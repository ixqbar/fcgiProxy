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
	Uuid string
	Conn *websocket.Conn
	JoinTime int64
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
		Conn:conn,
		JoinTime:time.Now().Unix(),
	}
}

func NewRequestClients() *RequestClients {
	return &RequestClients{
		Clients:make(map[string]*Client),
	}
}

func (obj *RequestClients) AddNewClient(uuid string, conn *websocket.Conn)  {
	obj.Lock()
	defer obj.Unlock()

	atomic.AddInt64(&obj.Num, 1)
	obj.Clients[uuid] = NewClient(uuid, conn)
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

	return obj.Clients[uuid].Conn.WriteMessage(websocket.TextMessage, message)
}

func (obj *RequestClients) BroadcastMessage(message []byte) error  {
	obj.Lock()
	defer obj.Unlock()

	for _, val := range obj.Clients {
		err := val.Conn.WriteMessage(websocket.TextMessage, message)
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