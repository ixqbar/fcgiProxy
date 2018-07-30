package proxy

import (
	"context"
	"errors"
	"github.com/jonnywang/go-kits/redis"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"github.com/google/uuid"
	"encoding/json"
)

var (
	ERR_PARAMS = errors.New("error params")
)

type FcgiRedisHandle struct {
	redis.RedisHandler
	sync.Mutex
}

func (obj *FcgiRedisHandle) Init() error {
	obj.Initiation(func() {
		GAndroidPushDevices = NewAndroidPushDevices()
		for _, device := range GConfig.ApushDevices {
			GAndroidPushDevices.AddDevice(device)
		}
	})

	return nil
}

func (obj *FcgiRedisHandle) Shutdown() {
	Logger.Print("redis server will shutdown")
}

func (obj *FcgiRedisHandle) Version() (string, error) {
	return VERSION, nil
}

func (obj *FcgiRedisHandle) Number() (int, error) {
	return Clients.Number(), nil
}

func (obj *FcgiRedisHandle) Uuid() (string, error) {
	return uuid.New().String(), nil
}

func (obj *FcgiRedisHandle) Del(clientUUID string) error {
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

func (obj *FcgiRedisHandle) Ping(message string) (string, error) {
	if len(message) > 0 {
		return message, nil
	}

	return "PONG", nil
}

func (obj *FcgiRedisHandle) Rpush(key string, content []byte) (error) {
	if len(key) == 0 || len(content) == 0 {
		return ERR_PARAMS
	}

	go AddNewProxyConfig(content)

	return nil
}

func (obj *FcgiRedisHandle) Qpush(group, message string) (error) {
	if len(group) == 0 || len(message) == 0 {
		return nil
	}

	go QpushMessage(group, message)

	return nil
}

func (obj *FcgiRedisHandle) Apush(group string, message []byte) (error) {
	if len(group) == 0 || len(message) == 0 {
		return nil
	}

	messageData := &TPushMessageData{}
	err := json.Unmarshal(message, messageData)
	if err != nil {
		return err
	}

	go GAndroidPushDevices.PushMessage(group, messageData)

	return nil
}

func (obj *FcgiRedisHandle) Tpush(group string, message []byte) (error) {
	if len(group) == 0 || len(message) == 0 {
		return nil
	}

	messageData := &TPushMessageData{}
	err := json.Unmarshal(message, messageData)
	if err != nil {
		return err
	}

	go func() {
		//iOS
		QpushMessage(group, messageData.Message)
		//android
		GAndroidPushDevices.PushMessage(group, messageData)
		//monitor
		if group == "*" {
			Clients.BroadcastMessage(NewClientTextMessage(message), MessageToMonitorClient)
		} else {
			Clients.PushMessage(group, NewClientTextMessage(message), MessageToMonitorClient)
		}
	}()

	return nil
}

func (obj *FcgiRedisHandle) Exists(clientUUID string) (int, error) {
	if Clients.GetClient(clientUUID) != nil {
		return 1, nil
	}

	return 0, nil
}

func (obj *FcgiRedisHandle) Npush(clientUUID string, message []byte) error {
	if clientUUID == "*" {
		Clients.BroadcastMessage(NewClientTextMessage(message), MessageToMonitorClient)
	} else {
		Clients.PushMessage(clientUUID, NewClientTextMessage(message), MessageToMonitorClient)
	}

	return nil
}

func (obj *FcgiRedisHandle) Setex(clientUUID string, messageType int, message []byte) error {
	return obj.Set(clientUUID, message, messageType)
}

func (obj *FcgiRedisHandle) Set(clientUUID string, message []byte, messageType int) error {
	var clientMessage *ClientMessage;
	if messageType == 0 {
		clientMessage = NewClientTextMessage(message)
	} else {
		clientMessage = NewClientBinaryMessage(message)
	}

	if clientUUID == "*" {
		Clients.BroadcastMessage(clientMessage, MessageToRequestClient)
	} else {
		Clients.PushMessage(clientUUID, clientMessage, MessageToRequestClient)
	}

	return nil
}

func (obj *FcgiRedisHandle) Subscribe(client *redis.Client, channelNames ...[]byte) (*redis.MultiChannelWriter, error) {
	if len(channelNames) == 0 {
		return nil, ERR_PARAMS
	}

	client.UseSubscribe = true
	client.Handler = &obj.RedisHandler

	ret := redis.NewMultiChannelWriter(len(channelNames))
	for _, channelName := range channelNames {
		cw := redis.NewChannelWriter(client.Host, string(channelName))
		if obj.SubChannels[string(channelName)] == nil {
			obj.SubChannels[string(channelName)] = []*redis.ChannelWriter{cw}
		} else {
			obj.SubChannels[string(channelName)] = append(obj.SubChannels[string(channelName)], cw)
		}
		ret.ChannelWriters = append(ret.ChannelWriters, cw)
	}

	return ret, nil
}

func (obj *FcgiRedisHandle) Publish(channelName string, message []byte) (int, error) {
	Logger.Printf("publish message to %s[%d:%d]", channelName, len(obj.SubChannels), len(obj.SubChannels[channelName]))
	if len(message) == 0 {
		return 0, nil
	}

	v, ok := obj.SubChannels[channelName]
	if !ok {
		return 0, nil
	}

	i := 0
	for _, c := range v {
		err := c.PublishMessage(message)
		if err != nil {
			Logger.Printf("publish %s to %s failed %s", channelName, c.ClientRequest.Host, err)
			continue
		}
		i++
	}

	return i, nil
}

var FcgiRedis = &FcgiRedisHandle{}

func Run() {
	err := FcgiRedis.Init()
	if err != nil {
		Logger.Print(err)
		return
	}

	server, err := redis.NewServer(GConfig.AdminServerAddress, FcgiRedis)
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
		FcgiRedis.Shutdown()
		server.Stop(10)
		err := httpServer.Shutdown(ctx)
		if err != nil {
			Logger.Print(err)
		}
		LoggerMessageRecord.Stop()
		redisStop <- 1
	}()

	Logger.Printf("redis protocol server run at %s", GConfig.AdminServerAddress)

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
