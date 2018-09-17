package proxy

import (
	"math/rand"
	"strings"
	"sync"
)

type FcgiServer struct {
	sync.Mutex
	address []string
}

func NewFcgiServer() *FcgiServer {
	return &FcgiServer{
		address: make([]string, 0),
	}
}

func (obj *FcgiServer) Init()  {
	obj.Lock()
	defer obj.Unlock()

	obj.address = strings.Split(GConfig.FcgiServerAddress, ",")
}

func (obj *FcgiServer) GetServer() string {
	obj.Lock()
	defer obj.Unlock()

	i := rand.Intn(len(obj.address))

	Logger.Printf("select fcgi server %s to post", obj.address[i])

	return obj.address[i]
}

func (obj *FcgiServer) AddServer(server string) bool {
	obj.Lock()
	defer obj.Unlock()

	for _, v := range obj.address {
		if v == server {
			return true
		}
	}

	obj.address = append(obj.address, server)

	return true
}

func (obj *FcgiServer) RemoveServer(server string) bool {
	obj.Lock()
	defer obj.Unlock()

	l := len(obj.address)

	for k, v := range obj.address {
		if v == server {
			if k != l-1 {
				obj.address[k] = obj.address[l-1]
			}
			obj.address = obj.address[:l-1]
			break
		}
	}

	return false
}

var GFcgiServer = NewFcgiServer()