package proxy

import (
	"errors"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"github.com/jonnywang/go-kits/redis"
	"context"
	"time"
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

}

func (obj *proxyRedisHandle) Version() (string, error) {
	return VERSION, nil
}

func (obj *proxyRedisHandle) Number() (int, error) {
	return Clients.Number(), nil
}

func (obj *proxyRedisHandle) Set(uuid string, message []byte) (error) {
	if uuid == "*" {
		Clients.BroadcastMessage(message)
	} else {
		Clients.PushMessage(uuid, message)
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

	sigs := make(chan os.Signal)
	signal.Notify(sigs, syscall.SIGTERM, syscall.SIGINT)

	ctx, cancel := context.WithCancel(context.Background())

	httpServer := WebSocket(ctx)

	go func() {
		<-sigs
		Logger.Print("catch exit signal")
		cancel()
		server.Stop(10)
		proxyRedisHandle.Shutdown()
		err := httpServer.Shutdown(ctx)
		if err != nil {
			Logger.Print(err)
		}
	}()

	Logger.Printf("redis protocol server run at %s", Config.AdminServerAddress)

	err = server.Start()
	if err != nil {
		Logger.Print(err)
	}

	time.Sleep(time.Duration(5) * time.Second)
}

