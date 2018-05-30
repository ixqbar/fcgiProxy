package proxy

import (
	"encoding/json"
)

type TProxyItem struct {
	Category string `json:"category"`
	Country  string `json:"country"`
	Address  string `json:"address"`
	Port     string `json:"port"`
	Time     int64  `json:"time"`
}

func AddNewProxyConfig(content []byte) {
	var proxyItem TProxyItem
	err := json.Unmarshal(content, &proxyItem)
	if err != nil {
		Logger.Print(err)
		return
	}

	Logger.Printf("add new proxy config %v", proxyItem)

	Config.ClearEmptyProxy()

	switch proxyItem.Category {
	case "socks5":
		Config.AddProxyConfig(TProxyIsSocks, proxyItem.Address, proxyItem.Port)
	case "http":
		Config.AddProxyConfig(TProxyIsHttp, proxyItem.Address, proxyItem.Port)
	}
}