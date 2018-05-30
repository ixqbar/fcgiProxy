package proxy

import (
	"encoding/json"
	"fmt"
)

type TProxyItem struct {
	Category string `json:"category"`
	Country  string `json:"country"`
	Address  string `json:"address"`
	Port     string `json:"port"`
	Time     int64  `json:"time"`
}

var ProxyPool = make(chan *TProxyConfig, MaxPoolSize)
var AvailableProxyPool = make(chan *TProxyConfig, MaxPoolSize)

func PushProxyPool(proxyConfig *TProxyConfig) {
	if len(ProxyPool) == MaxPoolSize {
		Logger.Printf("add proxy to full default pool fail %v", proxyConfig.String())
		return
	}

	Logger.Printf("will proxy to default pool %v,now default proxy pool size %d", proxyConfig.String(), len(ProxyPool))
	ProxyPool <- proxyConfig
	Logger.Printf("add proxy to default pool success %v,now default proxy pool size %d", proxyConfig.String(), len(ProxyPool))
}

func PushAvailableProxyPool(proxyConfig *TProxyConfig) {
	if len(AvailableProxyPool) == MaxPoolSize {
		Logger.Printf("add proxy to full available pool fail %v", proxyConfig.String())
		PushProxyPool(proxyConfig)
		return
	}

	Logger.Printf("will proxy to available pool %v,now available proxy pool size %d", proxyConfig.String(), len(AvailableProxyPool))
	AvailableProxyPool <- proxyConfig
	Logger.Printf("add proxy to available pool success %v,now available proxy pool size %d", proxyConfig.String(), len(AvailableProxyPool))
}

func PopProxyPool() *TProxyConfig {
	proxyConfig := PopAvailableProxyPool()
	if proxyConfig != nil {
		return proxyConfig
	}

	if len(ProxyPool) == 0 {
		for _, v := range Config.ProxyList {
			if v.Type != TProxyIsNone {
				PushProxyPool(&v)
			}
		}
	}

	select {
	case proxyConfig := <-ProxyPool:
		Logger.Printf("get proxy from default pool success %v,now default proxy pool size %d", proxyConfig.String(), len(ProxyPool))
		return proxyConfig
	default:
		Logger.Print("get proxy from default pool fail")
		return &TProxyConfig{TProxyIsNone, ""}
	}
}

func PopAvailableProxyPool() *TProxyConfig {
	select {
	case proxyConfig := <-AvailableProxyPool:
		Logger.Printf("get proxy from available pool success %v,now available proxy pool size %d", proxyConfig.String(), len(AvailableProxyPool))
		return proxyConfig
	default:
		return nil
	}
}

func AddNewProxyConfig(content []byte) {
	var proxyItem TProxyItem
	err := json.Unmarshal(content, &proxyItem)
	if err != nil {
		Logger.Print(err)
		return
	}

	Logger.Printf("add new proxy config %v", proxyItem)

	switch proxyItem.Category {
	case "socks5":
		PushProxyPool(&TProxyConfig{
			TProxyIsSocks,
			fmt.Sprintf("%s:%s", proxyItem.Address, proxyItem.Port),
		})
	case "http":
		PushProxyPool(&TProxyConfig{
			TProxyIsHttp,
			fmt.Sprintf("%s:%s", proxyItem.Address, proxyItem.Port),
		})
	}
}
