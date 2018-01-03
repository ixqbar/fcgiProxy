package proxy

import (
	"encoding/json"
	"time"
)

type PubSubMessage struct {
	UUID string `json:"uuid"`
	IP string 	`json:"ip"`
	Port string `json:"port"`
	Vars string `json:"vars"`
	Message string `json:"message"`
	Time time.Time `json:"time"`
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
	obj.Message = string(message)
	obj.Time = time.Now()
}

func (obj *PubSubMessage) Data() []byte {
	ret, err := json.Marshal(obj)
	if err != nil {
		Logger.Print(err)
	}

	return ret
}
