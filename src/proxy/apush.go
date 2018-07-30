package proxy

import (
	"encoding/xml"
	"errors"
	"fmt"
	"strings"
	"sync"
	"net/http"
	"io/ioutil"
	"encoding/json"
	"bytes"
)

type TPushMessageData struct {
	Title string `json:"title"`
	Message string `json:"message"`
}

func (obj TPushMessageData) String() string {
	return fmt.Sprintf("title=%s,message=%s", obj.Title, obj.Message)
}

type TApushMessage struct {
	To []string `json:"to"`
	Data *TPushMessageData `json:"data"`
}

type TApushDevice struct {
	group string
	name  string
	token string
}

func (obj *TApushDevice) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var content string
	if err := d.DecodeElement(&content, &start); err != nil {
		return err
	}

	if strings.Index(content, ",") >= 0 {
		deviceInfo := strings.Split(content, ",")
		obj.group = deviceInfo[0]
		obj.name = deviceInfo[1]
		obj.token = deviceInfo[2]
		return nil
	}

	return errors.New(fmt.Sprintf("error qpush device %s", content))
}

func (obj *TApushDevice) String() string {
	return fmt.Sprintf("group:%s,name:%s,token:%s", obj.group, obj.name, obj.token)
}

type TApushDevices struct {
	sync.Mutex
	devices map[string]*TApushDevice
}

func NewApushDevices() *TApushDevices {
	return &TApushDevices{
		devices: make(map[string]*TApushDevice, 0),
	}
}

func (obj *TApushDevices) AddDevice(device *TApushDevice) {
	obj.Lock()
	defer obj.Unlock()

	obj.devices[device.name] = device
}

func (obj *TApushDevices) UpdateDeviceToken(name, token string) bool {
	obj.Lock()
	defer obj.Unlock()

	_, ok := obj.devices[name]
	if !ok {
		return false
	}

	obj.devices[name].token = token

	return true
}

func (obj *TApushDevices) RemoveDevice(name string) {
	obj.Lock()
	defer obj.Unlock()

	delete(obj.devices, name)
}

func (obj *TApushDevices) PushMessage(group string, message *TPushMessageData) int {
	obj.Lock()
	defer obj.Unlock()

	num := 0
	if len(obj.devices) == 0 {
		return num
	}

	foundDevices := make([]string, 0)
	for _, device := range obj.devices {
		if group != "*" && device.group != group {
			continue
		}

		foundDevices = append(foundDevices, device.token)
	}

	num = len(foundDevices)
	if num > 0 {
		go doApushMessage(foundDevices, message)
	}

	return num
}

func doApushMessage(devices []string, message *TPushMessageData) {
	bodyData, err := json.Marshal(TApushMessage{To:devices,Data:message})
	if err != nil {
		Logger.Print(err)
		return
	}

	Logger.Print(string(bodyData))

	request, err := http.NewRequest("POST", GConfig.ApushUrl, bytes.NewBuffer(bodyData))
	if err != nil {
		Logger.Print(err)
		return
	}

	request.Header.Set("Content-Type", "application/json")
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		Logger.Print(err)
		return
	}

	defer response.Body.Close()

	content, err := ioutil.ReadAll(response.Body)
	if err != nil {
		Logger.Print(err)
		return
	}

	Logger.Printf("send message `%s` to android devices %v got response `%s`", message, devices, string(content))
}
