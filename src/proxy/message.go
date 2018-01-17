package proxy

import (
	"encoding/json"
	"time"
	"github.com/google/uuid"
)

const (
	PubSubMessageTypeIsProxy = "proxy"
	PubSubMessageTypeIsLogs  = "logs"
)

type PubSubMessage struct {
	Type string `json:"type"`
	ID string `json:"id"`
	UUID string `json:"uuid"`
	IP string 	`json:"ip"`
	Port string `json:"port"`
	Vars string `json:"vars"`
	UserAgent string `json:"user_agent"`
	Message interface{} `json:"message"`
	Time int64 `json:"time"`
	LogTryNum int `json:"log_try_num"`
}

func NewPubSubMessage(uuid, ip, port, vars, agent string) *PubSubMessage {
	return &PubSubMessage{
		UUID:uuid,
		IP:ip,
		Port:port,
		Vars:vars,
		UserAgent:agent,
		LogTryNum:0,
	}
}

func (obj *PubSubMessage) UpdateMessage(messageType string, messageContent interface{})  {
	obj.Type = messageType
	obj.ID = uuid.New().String()
	obj.Message = messageContent
	obj.Time = time.Now().UnixNano() / 1e6
}

func (obj *PubSubMessage) Data() []byte {
	ret, err := json.Marshal(obj)
	if err != nil {
		Logger.Print(err)
	}

	return ret
}

func (obj *PubSubMessage) Durable() {
	LoggerMessageRecord.RecordMessage(obj)
}
