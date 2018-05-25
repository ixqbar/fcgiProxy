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
	"net/url"
)

type ClientMessage struct {
	category int
	data     []byte
}

type Client struct {
	sync.Mutex
	UUID          string
	conn          *websocket.Conn
	joinTime      int64
	alive         bool
	over          chan bool
	request       *http.Request
	requestValues *url.Values
	message       chan *ClientMessage
}

func NewClientTextMessage(message []byte) *ClientMessage {
	return &ClientMessage{
		category: websocket.TextMessage,
		data:     message,
	}
}

func NewClientBinaryMessage(message []byte) *ClientMessage {
	return &ClientMessage{
		category: websocket.BinaryMessage,
		data:     message,
	}
}

func NewClientMessage(category int, message []byte) *ClientMessage {
	return &ClientMessage{
		category: category,
		data:     message,
	}
}

func (obj *Client) PipeSendMessage() {
	ticker := time.NewTicker(15 * time.Second)

	obj.conn.SetPongHandler(func(appData string) error {
		Logger.Printf("client %s[%s] got pong message %s", obj.conn.RemoteAddr(), obj.UUID, appData)
		return obj.conn.SetReadDeadline(time.Now().Add(30 * time.Second))
	})

	defer ticker.Stop()

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

			if err := obj.conn.WriteMessage(message.category, message.data); err != nil {
				Logger.Print(err)
				return
			}
		}
	}

	Logger.Printf("client %s[%s] send message end", obj.conn.RemoteAddr(), obj.UUID)
}

func (obj *Client) PipeReadMessage() {
	qstr := Config.QueryString
	if len(qstr) > 0 {
		if len(obj.request.URL.RawQuery) > 0 {
			qstr = fmt.Sprintf("%s&%s", qstr, obj.request.URL.RawQuery)
		}
	} else {
		qstr = obj.request.URL.RawQuery
	}

	Logger.Printf("client %s[%s] final query[%s]", obj.conn.RemoteAddr(), obj.UUID, qstr)

	var env map[string]string
	var body *bytes.Reader

	remoteInfo := strings.Split(obj.conn.RemoteAddr().String(), ":")

	if len(Config.ScriptFileName) > 0 {
		env = make(map[string]string)
		env["SCRIPT_FILENAME"] = Config.ScriptFileName
		env["QUERY_STRING"] = qstr

		for _, item := range Config.HeaderParams {
			env[item.Key] = item.Value
		}

		env["REMOTE_ADDR"] = remoteInfo[0]
		env["REMOTE_PORT"] = remoteInfo[1]
		env["PROXY_UUID"] = obj.UUID

		body = bytes.NewReader(nil)
	}

	requestNoProxy := IsFalse(obj.requestValues.Get("proxy"))
	pubSubChannel := obj.requestValues.Get("channel")
	pubSubMessage := NewPubSubMessage(obj.UUID, remoteInfo[0], remoteInfo[1], qstr, obj.request.Header.Get("User-Agent"))

	var pubSubData []byte

	frameMessageType := websocket.TextMessage
	if len(Config.HttpRc4EncryptKey) > 0 {
		frameMessageType = websocket.BinaryMessage
	}

	for {
		messageType, messageContent, err := obj.conn.ReadMessage()
		if err != nil {
			Logger.Printf("client %s[%s] read err message failed %s", obj.conn.RemoteAddr(), obj.UUID, err)
			break
		}

		if messageType != frameMessageType {
			Logger.Printf("client %s[%s] read err message type", obj.conn.RemoteAddr(), obj.UUID)
			break
		}

		if messageType == websocket.BinaryMessage {
			if FcgiRedis.CanPublish() {
				originMessage, err := Rc4Decrypt(messageContent, []byte(Config.HttpRc4EncryptKey))
				if err != nil {
					Logger.Printf("client %s[%s] decrypt message fail", obj.conn.RemoteAddr(), obj.UUID)
					break
				}

				pubSubMessage.UpdateMessage(PubSubMessageTypeIsProxy, string(originMessage))
				pubSubData = pubSubMessage.Data()

				FcgiRedis.Publish("*", pubSubData)
				if len(pubSubChannel) > 0 {
					FcgiRedis.Publish(pubSubChannel, pubSubData)
				}
			}
		} else if FcgiRedis.CanPublish() {
			pubSubMessage.UpdateMessage(PubSubMessageTypeIsProxy, string(messageContent))
			pubSubData = pubSubMessage.Data()

			FcgiRedis.Publish("*", pubSubData)
			if len(pubSubChannel) > 0 {
				FcgiRedis.Publish(pubSubChannel, pubSubData)
			}
		}

		if body == nil || requestNoProxy == true {
			continue
		}

		body.Reset(messageContent)

		startTime := time.Now()

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

		err = obj.PushMessage(NewClientMessage(frameMessageType, content))
		if err != nil {
			Logger.Printf("client %s[%s] response failed %s", obj.conn.RemoteAddr(), obj.UUID, err)
			break
		}

		Logger.Printf("client %s[%s] request success cost time %s", obj.conn.RemoteAddr(), obj.UUID, time.Since(startTime).String())
	}
}

func (obj *Client) PushMessage(clientMessage *ClientMessage) error {
	select {
	case obj.message <- clientMessage:
	default:
		return errors.New(fmt.Sprintf("push message to client %s[%s] failed", obj.conn.RemoteAddr(), obj.UUID))
	}

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
	close(obj.message)
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

func NewClient(uuid string, conn *websocket.Conn, r *http.Request, rv *url.Values) *Client {
	return &Client{
		UUID:          uuid,
		conn:          conn,
		joinTime:      time.Now().Unix(),
		over:          make(chan bool),
		alive:         true,
		request:       r,
		requestValues: rv,
		message:       make(chan *ClientMessage),
	}
}

func NewRequestClients() *RequestClients {
	return &RequestClients{
		Clients: make(map[string]*Client),
	}
}

func (obj *RequestClients) AddNewClient(uuid string, conn *websocket.Conn, r *http.Request, rv *url.Values) *Client {
	obj.Lock()
	defer obj.Unlock()

	obj.num++
	obj.Clients[uuid] = NewClient(uuid, conn, r, rv)

	return obj.Clients[uuid]
}

func (obj *RequestClients) RemoveClient(uuid string) {
	obj.Lock()
	defer obj.Unlock()

	obj.Clients[uuid].Remove()
	obj.num--
	delete(obj.Clients, uuid)
}

func (obj *RequestClients) PushMessage(uuid string, clientMessage *ClientMessage) error {
	obj.Lock()
	defer obj.Unlock()

	if _, ok := obj.Clients[uuid]; !ok {
		return errors.New(fmt.Sprintf("not found client %s", uuid))
	}

	return obj.Clients[uuid].PushMessage(clientMessage)
}

func (obj *RequestClients) BroadcastMessage(clientMessage *ClientMessage) error {
	obj.Lock()
	defer obj.Unlock()

	for _, val := range obj.Clients {
		err := val.PushMessage(clientMessage)
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
