package proxy

import (
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"sync"
	"time"
	"net/http"
	"github.com/tomasen/fcgi_client"
	"io/ioutil"
	"strings"
	"bytes"
)

type Client struct {
	sync.Mutex
	UUID     string
	conn     *websocket.Conn
	joinTime int64
	alive    bool
	over     chan bool
	request  *http.Request
	message chan []byte
}

func (obj *Client) PipeSendMessage() {
	ticker := time.NewTicker(30 * time.Second)

	Logger.Printf("client %s[%s] waiting for message send", obj.conn.RemoteAddr(), obj.UUID)

	obj.conn.SetPongHandler(func(appData string) error {
		Logger.Printf("client %s[%s] got pong message %s", obj.conn.RemoteAddr(), obj.UUID, appData)
		return obj.conn.SetReadDeadline(time.Now().Add(30 * time.Second))
	})

	for {
		select {
			case <-ticker.C:
				if err := obj.conn.WriteMessage(websocket.PingMessage, []byte("PING")); err != nil {
					return
				}
				Logger.Printf("client %s[%s] send ping message PING", obj.conn.RemoteAddr(), obj.UUID)
			case message, ok := <-obj.message:
				if !ok {
					obj.conn.WriteMessage(websocket.CloseMessage, []byte{})
					return
				}

				if err := obj.conn.WriteMessage(websocket.TextMessage, message); err != nil {
					Logger.Print(err)
					return
				}
		}
	}

	Logger.Printf("client %s[%s] send message end", obj.conn.RemoteAddr(), obj.UUID)
}

func (obj *Client) PipeReadMessage() {
	qstr := obj.request.URL.RawQuery
	if len(qstr) == 0 {
		qstr = Config.QueryString
	} else {
		qstr = fmt.Sprintf("%s&%s", qstr, Config.QueryString)
	}

	Logger.Printf("client %s[%s] final query[%s]", obj.conn.RemoteAddr(), obj.UUID, qstr)

	env := make(map[string]string)
	env["SCRIPT_FILENAME"] = Config.ScriptFileName
	env["QUERY_STRING"] = qstr

	for _, item := range Config.HeaderParams {
		env[item.Key] = item.Value
	}

	remoteInfo := strings.Split(obj.conn.RemoteAddr().String(), ":")
	env["REMOTE_ADDR"] = remoteInfo[0]
	env["REMOTE_PORT"] = remoteInfo[1]
	env["PROXY_UUID"] = obj.UUID

	body := bytes.NewReader(nil)

	for {
		messageType, p, err := obj.conn.ReadMessage()
		if err != nil {
			Logger.Printf("client %s[%s] read err message failed %s", obj.conn.RemoteAddr(), obj.UUID, err)
			break
		}

		if messageType != websocket.TextMessage {
			Logger.Printf("client %s[%s] read err message type", obj.conn.RemoteAddr(), obj.UUID)
			break
		}

		body.Reset(p)

		Logger.Printf("client %s[%s] request body [%s][%d]", obj.conn.RemoteAddr(), obj.UUID, string(p), body.Len())

		fcgi, err := fcgiclient.Dial("tcp", Config.FcgiServerAddress)
		if err != nil {
			Logger.Print(err)
			break
		}

		resp, err := fcgi.Post(env, "application/octet-stream", body, body.Len())
		if err != nil {
			Logger.Printf("client %s[%s] read fcgi response failed %s", obj.conn.RemoteAddr(), obj.UUID, err)
			fcgi.Close()
			break
		}

		content, err := ioutil.ReadAll(resp.Body)
		fcgi.Close()

		if err != nil {
			Logger.Printf("client %s[%s] read fcgi response failed %s", obj.conn.RemoteAddr(), obj.UUID, err)
			break
		}

		err = obj.PushMessage(content)
		if err != nil {
			Logger.Printf("client %s[%s] response failed %s", obj.conn.RemoteAddr(), obj.UUID, err)
			break
		}
	}
}

func (obj *Client) PushMessage(message []byte) error {
	obj.message <- message

	return nil
}

func (obj *Client) Close() {
	obj.Lock()
	defer obj.Unlock()

	err := obj.conn.SetReadDeadline(time.Now())
	if err != nil {
		Logger.Printf("client %s[%s] SetReadDeadline failed %s", obj.conn.RemoteAddr(), obj.UUID, err)
	}
	obj.alive = false
	<-obj.over
}

func (obj *Client) Remove() {
	if obj.alive {
		go obj.Close()
	}

	obj.over <- true
}

type RequestClients struct {
	sync.Mutex
	num     int
	Clients map[string]*Client
}

var Clients = NewRequestClients()

func NewClient(uuid string, conn *websocket.Conn, r *http.Request) *Client {
	return &Client{
		UUID:     uuid,
		conn:     conn,
		joinTime: time.Now().Unix(),
		over:     make(chan bool),
		alive:    true,
		request:  r,
		message: make(chan[]byte),
	}
}

func NewRequestClients() *RequestClients {
	return &RequestClients{
		Clients: make(map[string]*Client),
	}
}

func (obj *RequestClients) AddNewClient(uuid string, conn *websocket.Conn, r *http.Request) *Client {
	obj.Lock()
	defer obj.Unlock()

	obj.num++
	obj.Clients[uuid] = NewClient(uuid, conn, r)

	return obj.Clients[uuid]
}

func (obj *RequestClients) RemoveClient(uuid string) {
	obj.Lock()
	defer obj.Unlock()

	obj.Clients[uuid].Remove()
	obj.num--
	delete(obj.Clients, uuid)
}

func (obj *RequestClients) PushMessage(uuid string, message []byte) error {
	obj.Lock()
	defer obj.Unlock()

	if _, ok := obj.Clients[uuid]; !ok {
		return errors.New(fmt.Sprintf("not found client %s", uuid))
	}

	return obj.Clients[uuid].PushMessage(message)
}

func (obj *RequestClients) BroadcastMessage(message []byte) error {
	obj.Lock()
	defer obj.Unlock()

	for _, val := range obj.Clients {
		err := val.PushMessage(message)
		if err != nil {
			Logger.Printf("broadcast message to %s failed %s", val.UUID, err)
		}
	}

	return nil
}

func (obj *RequestClients) Number() int {
	obj.Lock()
	defer obj.Unlock()

	return obj.num
}

func (obj *RequestClients) GetClient(uuid string) *Client {
	obj.Lock()
	defer obj.Unlock()

	if _, ok := obj.Clients[uuid]; !ok {
		return nil
	}

	return obj.Clients[uuid]
}

func (obj *RequestClients) RemoveAll() {
	for key, val := range obj.Clients {
		val.Close()
		delete(obj.Clients, key)
	}
}
