package proxy

import (
	"net/http"
	"fmt"
	"strings"
	"io/ioutil"
	"encoding/json"
)

type TQpushSignResponse struct {
	Error string `json:"error"`
	Data string `json:"pusherData"`
	Sign string `json:"pusheeSig"`
}

func CheckDeviceSign() error {
	if len(Config.QpushDevices) == 0 {
		Logger.Print("not found qpush devices config")
		return nil
	}

	httpClient, err := MakeHttpClient()
	if err != nil {
		Logger.Print(err)
		return err
	}

	url := "https://qpush.me/pusher/checkphone/"
	qPushSignResponse := TQpushSignResponse{}

	for deviceIndex, deviceInfo := range Config.QpushDevices {
		data := fmt.Sprintf("pusher_id=71455&type=chromeExt&version_client=66.0.3359.170&version_app=1.4&name=%s&code=%s", deviceInfo.Name, deviceInfo.Code)
		request, err := http.NewRequest("POST", url, strings.NewReader(data))
		if err != nil {
			Logger.Print(err)
			return err
		}

		request.Header.Set("Content-Type", "application/x-www-form-urlencoded");
		request.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_12_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/66.0.3359.170 Safari/537.36");
		response, err := httpClient.Do(request)
		if err != nil {
			Logger.Print(err)
			return err
		}

		responseBody, err := ioutil.ReadAll(response.Body)
		if err != nil {
			Logger.Print(err)
			return err
		}

		Logger.Printf("qpush got sign resposne %s", responseBody)

		err = json.Unmarshal(responseBody, &qPushSignResponse)
		if err != nil {
			Logger.Print(err)
			return err
		}

		if len(qPushSignResponse.Sign) > 0 {
			Config.QpushDevices[deviceIndex].Sign = qPushSignResponse.Sign
			Config.QpushDevices[deviceIndex].Available = true
			Logger.Printf("qpush got sign %s success for %s[%s][%s]", qPushSignResponse.Sign, deviceInfo.Name, deviceInfo.Group, deviceInfo.Code)
		} else {
			Logger.Printf("qpush got sign %v failed for %s[%s][%s]", qPushSignResponse, deviceInfo.Name, deviceInfo.Group, deviceInfo.Code)
		}
	}

	return nil
}

func QpushMessage(group, message string)  {
	if len(Config.QpushDevices) == 0 {
		return
	}

	httpClient, err := MakeHttpClient()
	if err != nil {
		Logger.Print(err)
		return
	}

	url := "https://qpush.me/pusher/push_site/"
	for _, deviceInfo := range Config.QpushDevices {
		if deviceInfo.Available == false || deviceInfo.Group != group {
			continue
		}

		data := fmt.Sprintf("name=%s&code=%s&sig=%s&cache=true&msg[text]=%s", deviceInfo.Name, deviceInfo.Code, deviceInfo.Sign, message)
		request, err := http.NewRequest("POST", url, strings.NewReader(data))
		if err != nil {
			Logger.Print(err)
			return
		}

		request.Header.Set("Content-Type", "application/x-www-form-urlencoded");
		request.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_12_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/66.0.3359.170 Safari/537.36");
		response, err := httpClient.Do(request)
		if err != nil {
			Logger.Print(err)
		}

		responseBody, err := ioutil.ReadAll(response.Body)
		if err != nil {
			Logger.Print(err)
		}

		Logger.Printf("qpush message %s to %s[%s][%s][%s] response %s", message, deviceInfo.Name, deviceInfo.Group, deviceInfo.Code, deviceInfo.Sign, responseBody)
	}
}