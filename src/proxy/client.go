package proxy

import (
	"github.com/gorilla/websocket"
	"time"
	"sync"
	"sync/atomic"
	"errors"
	"fmt"
)

type Client struct {
	sync.Mutex
	Uuid string
	conn *websocket.Conn
	joinTime int64
}

func (obj *Client) PushMessage(message []byte) error {
	obj.Lock()
	defer obj.Unlock()

	return obj.conn.WriteMessage(websocket.TextMessage, message)
}

type RequestClients struct {
	sync.Mutex
	Num int64
	Clients map[string]*Client
}

var Clients = NewRequestClients()

func NewClient(uuid string, conn *websocket.Conn) *Client {
	return &Client{
		Uuid:uuid,
		conn:conn,
		joinTime:time.Now().Unix(),
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

	atomic.AddInt64(&obj.Num, 1)
	obj.Clients[uuid] = NewClient(uuid, conn)

	return obj.Clients[uuid]
}

func (obj *RequestClients) RemoveClient(uuid string)  {
	obj.Lock()
	defer obj.Unlock()

	atomic.AddInt64(&obj.Num, -1)
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
			Logger.Printf("broadcast message to %s failed %s", val.Uuid, err)
		}
	}

	return nil
}

func (obj *RequestClients) Number() int64 {
	obj.Lock()
	defer obj.Unlock()

	return obj.Num
}