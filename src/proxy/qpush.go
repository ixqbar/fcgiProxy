package proxy

import (
	"net/http"
	"fmt"
	"strings"
	"io/ioutil"
	"encoding/json"
	"errors"
)

type TQpushSignResponse struct {
	Error string `json:"error"`
	Data  string `json:"pusherData"`
	Sign  string `json:"pusheeSig"`
}

func CheckDeviceSign(deviceIndex int) error {
	deviceInfo := Config.QpushDevices[deviceIndex]
	maxLoop := 5
	currentLoop := 0
	url := "https://qpush.me/pusher/checkphone/"
	qPushSignResponse := TQpushSignResponse{}

	data := fmt.Sprintf("pusher_id=71455&type=chromeExt&version_client=66.0.3359.170&version_app=1.4&name=%s&code=%s", deviceInfo.Name, deviceInfo.Code)
	request, err := http.NewRequest("POST", url, strings.NewReader(data))
	if err != nil {
		Logger.Print(err)
		return err
	}

	request.Header.Set("Content-Type", "application/x-www-form-urlencoded");
	request.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_12_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/66.0.3359.170 Safari/537.36");
	for currentLoop = 0; currentLoop < maxLoop; currentLoop++ {
		client, err := MakeHttpClient(Config.ProxyConfigIndex)
		if err != nil {
			Logger.Print(err)
			continue
		}

		response, err := client.httpClient.Do(request)
		if err != nil {
			Logger.Print(err)
			Config.RemoveProxyConfig(client.proxyIndex)
			continue
		}

		responseBody, err := ioutil.ReadAll(response.Body)
		if err != nil {
			Logger.Print(err)
			Config.RemoveProxyConfig(client.proxyIndex)
			continue
		}

		Logger.Printf("qpush got sign resposne %s", responseBody)

		err = json.Unmarshal(responseBody, &qPushSignResponse)
		if err != nil {
			Logger.Print(err)
			Config.RemoveProxyConfig(client.proxyIndex)
			continue
		}

		if len(qPushSignResponse.Sign) > 0 {
			Config.QpushDevices[deviceIndex].Sign = qPushSignResponse.Sign
			Config.QpushDevices[deviceIndex].Available = true
			Config.ProxyConfigIndex = client.proxyIndex
			Logger.Printf("qpush got sign %s success for %s[%s][%s]", qPushSignResponse.Sign, deviceInfo.Name, deviceInfo.Group, deviceInfo.Code)
			return nil
		} else {
			Config.RemoveProxyConfig(client.proxyIndex)
		}
	}

	Logger.Printf("qpush got sign %v failed for %s[%s][%s]", qPushSignResponse, deviceInfo.Name, deviceInfo.Group, deviceInfo.Code)

	return errors.New("qpush got sign fail")
}

func QpushMessage(group, message string) {
	if len(Config.QpushDevices) == 0 {
		return
	}

	maxLoop := 5
	currentLoop := 0
	url := "https://qpush.me/pusher/push_site/"
	for deviceIndex, deviceInfo := range Config.QpushDevices {
		if deviceInfo.Group != group {
			continue
		}

		if deviceInfo.Available == false {
			err := CheckDeviceSign(deviceIndex)
			if err != nil {
				continue
			}
		}

		data := fmt.Sprintf("name=%s&code=%s&sig=%s&cache=true&msg[text]=%s", deviceInfo.Name, deviceInfo.Code, deviceInfo.Sign, message)
		request, err := http.NewRequest("POST", url, strings.NewReader(data))
		if err != nil {
			Logger.Print(err)
			return
		}

		request.Header.Set("Content-Type", "application/x-www-form-urlencoded");
		request.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_12_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/66.0.3359.170 Safari/537.36");
		for currentLoop = 0; currentLoop < maxLoop; currentLoop++ {
			client, err := MakeHttpClient(Config.ProxyConfigIndex)
			if err != nil {
				Logger.Print(err)
				continue
			}

			response, err := client.httpClient.Do(request)
			if err != nil {
				Logger.Print(err)
				Config.RemoveProxyConfig(client.proxyIndex)
				continue
			}

			responseBody, err := ioutil.ReadAll(response.Body)
			if err != nil {
				Logger.Print(err)
				Config.RemoveProxyConfig(client.proxyIndex)
				continue
			}

			Config.ProxyConfigIndex = client.proxyIndex
			Logger.Printf("qpush message %s to %s[%s][%s][%s] response %s", message, deviceInfo.Name, deviceInfo.Group, deviceInfo.Code, deviceInfo.Sign, responseBody)
			break
		}
	}
}
