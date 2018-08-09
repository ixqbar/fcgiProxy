package proxy

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

type TQpushSignResponse struct {
	Error string `json:"error"`
	Data  string `json:"pusherData"`
	Sign  string `json:"pusheeSig"`
}

func CheckDeviceSign(deviceIndex int) error {
	deviceInfo := GConfig.QpushDevices[deviceIndex]

	url := "https://qpush.me/pusher/checkphone/"
	qPushSignResponse := TQpushSignResponse{}

	data := fmt.Sprintf("pusher_id=71455&type=chromeExt&version_client=66.0.3359.170&version_app=1.4&name=%s&code=%s", deviceInfo.Name, deviceInfo.Code)
	request, err := http.NewRequest("POST", url, strings.NewReader(data))
	if err != nil {
		Logger.Print(err)
		return err
	}

	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_12_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/66.0.3359.170 Safari/537.36")
	for tryNum := 0; tryNum < MaxQpushTryNum; tryNum++ {
		client, err := MakeHttpClient()
		if err != nil {
			Logger.Print(err)
			continue
		}

		response, err := client.httpClient.Do(request)
		if err != nil {
			Logger.Print(err)
			continue
		}

		responseBody, err := ioutil.ReadAll(response.Body)
		if err != nil {
			Logger.Print(err)
			continue
		}

		Logger.Printf("qpush got sign resposne %s", responseBody)

		err = json.Unmarshal(responseBody, &qPushSignResponse)
		if err != nil {
			Logger.Print(err)
			continue
		}

		if len(qPushSignResponse.Sign) == 0 {
			continue
		}

		client.Success()
		GConfig.QpushDevices[deviceIndex].Sign = qPushSignResponse.Sign
		GConfig.QpushDevices[deviceIndex].Available = true
		Logger.Printf("qpush got sign %s success for %s[%s][%s] by proxyServer %s",
			qPushSignResponse.Sign, deviceInfo.Name, deviceInfo.Group, deviceInfo.Code,
			client.proxyConfig.String(),
		)

		return nil
	}

	Logger.Printf("qpush got sign %v failed for %s[%s][%s]", qPushSignResponse, deviceInfo.Name, deviceInfo.Group, deviceInfo.Code)

	return errors.New("qpush got sign fail")
}

func QpushMessage(group, message string) {
	if len(GConfig.QpushDevices) == 0 {
		return
	}

	url := "https://qpush.me/pusher/push_site/"
	for deviceIndex, deviceInfo := range GConfig.QpushDevices {
		if group != "*" && deviceInfo.Group != group {
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

		request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		request.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_12_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/66.0.3359.170 Safari/537.36")
		for tryNum := 0; tryNum < MaxQpushTryNum; tryNum++ {
			client, err := MakeHttpClient()
			if err != nil {
				Logger.Print(err)
				continue
			}

			response, err := client.httpClient.Do(request)
			if err != nil {
				Logger.Print(err)
				continue
			}

			responseBody, err := ioutil.ReadAll(response.Body)
			if err != nil {
				Logger.Print(err)
				continue
			}

			client.Success()
			Logger.Printf("qpush message %s to %s[%s][%s][%s] response %s by proxyServer %s",
				message, deviceInfo.Name, deviceInfo.Group, deviceInfo.Code, deviceInfo.Sign, responseBody,
				client.proxyConfig.String(),
			)

			break
		}
	}
}
