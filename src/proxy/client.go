package proxy

import (
	"github.com/gorilla/websocket"
	"time"
	"sync"
	"errors"
	"fmt"
)

type Client struct {
	sync.Mutex
	UUID string
	conn *websocket.Conn
	joinTime int64
	alive bool
	over chan bool
}

func (obj *Client) PushMessage(message []byte) error {
	obj.Lock()
	defer obj.Unlock()

	return obj.conn.WriteMessage(websocket.TextMessage, message)
}

func (obj *Client) Close() {
	obj.Lock()
	defer obj.Unlock()

	obj.conn.SetReadDeadline(time.Now())
	obj.alive = false
	<- obj.over
}

func (obj *Client) remove() {
	if obj.alive {
		obj.Close()
	}

	obj.over <- true
}


type RequestClients struct {
	sync.Mutex
	num int
	Clients map[string]*Client
}

var Clients = NewRequestClients()

func NewClient(uuid string, conn *websocket.Conn) *Client {
	return &Client{
		UUID:uuid,
		conn:conn,
		joinTime:time.Now().Unix(),
		over:make(chan bool),
		alive:true,
	}
}

func NewRequestClients() *RequestClients {
	return &RequestClients{
		Clients:make(map[string]*Client),
	}
}

func (obj *RequestClients) AddNewClient(uuid string, conn *websocket.Conn) *Client {
	obj.Lock()
	defer obj.Unlock()

	obj.num++
	obj.Clients[uuid] = NewClient(uuid, conn)

	return obj.Clients[uuid]
}

func (obj *RequestClients) RemoveClient(uuid string)  {
	obj.Lock()
	defer obj.Unlock()

	obj.Clients[uuid].remove()
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

func (obj *RequestClients) BroadcastMessage(message []byte) error  {
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

func (obj *RequestClients) GetClient(uuid string) (*Client) {
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