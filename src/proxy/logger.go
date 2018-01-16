package proxy

import (
	"github.com/jonnywang/go-kits/redis"
	"log"
)

type LogMessage struct {
	UserID int `json:"id"`
	Resource string `json:"res"`
	Type string `json:"type"`
	Content string `json:"data"`
}

type LogMessageRecord struct {
	stopSignal chan int
	message chan *PubSubMessage
}

func NewLogMessageRecord() *LogMessageRecord  {
	return &LogMessageRecord{
		stopSignal:make(chan int),
		message:make(chan *PubSubMessage, 1024),
	}
}

func (obj *LogMessageRecord) Run() {
	go func() {
		defer LoggerMessageDao().Close()

		for {
			select {
			case <- obj.stopSignal:
				return
			case pubSubMessage,ok := <- obj.message:
				if !ok  {
					continue
				}

				Logger.Printf("record log messages %s", pubSubMessage.Data())
				LoggerMessageDao().RecordMessage(pubSubMessage)
			}
		}
	}()
}

func (obj *LogMessageRecord) RecordMessage(pbMessage *PubSubMessage) {
	obj.message <- pbMessage
}

func (obj *LogMessageRecord) Stop() {
	obj.stopSignal <- 1
}

var Logger = redis.Logger
var LoggerMessageRecord = NewLogMessageRecord()

func init() {
	redis.Logger.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
}
