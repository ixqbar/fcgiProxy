package proxy

import (
	"context"
	"errors"
	"github.com/jonnywang/go-kits/redis"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

var (
	ERR_PARAMS = errors.New("error params")
)

type proxyRedisHandle struct {
	redis.RedisHandler
	sync.Mutex
}

func (obj *proxyRedisHandle) Init() error {

	return nil
}

func (obj *proxyRedisHandle) Shutdown() {
	Logger.Print("redis server will shutdown")
}

func (obj *proxyRedisHandle) Version() (string, error) {
	return VERSION, nil
}

func (obj *proxyRedisHandle) Number() (int, error) {
	return Clients.Number(), nil
}

func (obj *proxyRedisHandle) Del(clientUUID string) error {
	if clientUUID == "*" {
		Clients.RemoveAll()
	} else {
		client := Clients.GetClient(clientUUID)
		if client != nil {
			client.Close()
		}
	}

	return nil
}

func (obj *proxyRedisHandle) Set(clientUUID string, message []byte) error {
	if clientUUID == "*" {
		Clients.BroadcastMessage(message)
	} else {
		Clients.PushMessage(clientUUID, message)
	}
	return nil
}

func Run() {
	proxyRedisHandle := &proxyRedisHandle{}

	proxyRedisHandle.SetShield("Init")
	proxyRedisHandle.SetShield("Shutdown")
	proxyRedisHandle.SetShield("Lock")
	proxyRedisHandle.SetShield("Unlock")
	proxyRedisHandle.SetShield("SetShield")
	proxyRedisHandle.SetShield("SetConfig")
	proxyRedisHandle.SetShield("CheckShield")

	err := proxyRedisHandle.Init()
	if err != nil {
		Logger.Print(err)
		return
	}

	server, err := redis.NewServer(Config.AdminServerAddress, proxyRedisHandle)
	if err != nil {
		Logger.Print(err)
		return
	}

	redisStop := make(chan int)
	stopSignal := make(chan os.Signal)
	signal.Notify(stopSignal, syscall.SIGTERM, syscall.SIGINT)

	ctx, cancel := context.WithCancel(context.Background())
	httpServer, httpStop := NewWebSocket()

	go func() {
		<-stopSignal
		Logger.Print("catch exit signal")
		cancel()
		proxyRedisHandle.Shutdown()
		server.Stop(10)
		err := httpServer.Shutdown(ctx)
		if err != nil {
			Logger.Print(err)
		}
		redisStop <- 1
	}()

	Logger.Printf("redis protocol server run at %s", Config.AdminServerAddress)

	err = server.Start()
	if err != nil {
		Logger.Print(err)
	}

	<-redisStop
	<-httpStop

	close(stopSignal)
	close(redisStop)
	close(httpStop)

	Logger.Print("all server shutdown")
}
