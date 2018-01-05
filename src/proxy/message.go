package proxy

import (
	"encoding/json"
	"time"
	"github.com/google/uuid"
)

type PubSubMessage struct {
	ID string `json:"id"`
	UUID string `json:"uuid"`
	IP string 	`json:"ip"`
	Port string `json:"port"`
	Vars string `json:"vars"`
	Message string `json:"message"`
	Time int64 `json:"time"`
}

func NewPubSubMessage(uuid string, ip string, port string, vars string) *PubSubMessage {
	return &PubSubMessage{
		UUID:uuid,
		IP:ip,
		Port:port,
		Vars:vars,
	}
}

func (obj *PubSubMessage) UpdateMessage(message []byte)  {
	obj.ID = uuid.New().String()
	obj.Message = string(message)
	obj.Time = time.Now().UnixNano() / 1e6
}

func (obj *PubSubMessage) Data() []byte {
	ret, err := json.Marshal(obj)
	if err != nil {
		Logger.Print(err)
	}

	return ret
}
