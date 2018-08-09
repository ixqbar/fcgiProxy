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

var DefaultProxyPool = make(chan *TProxyConfig, MaxPoolSize)
var AvailableProxyPool = make(chan *TProxyConfig, MaxPoolSize)

func AddProxyConfigToDefaultProxyPool(proxyConfig *TProxyConfig) {
	if len(DefaultProxyPool) == MaxPoolSize {
		Logger.Printf("add proxy to full default pool %v", proxyConfig.String())
		tmpProxyConf := <-DefaultProxyPool
		Logger.Printf("discard proxy from default pool %v", tmpProxyConf.String())
	}

	Logger.Printf("will proxy to default pool %v,now default proxy pool size %d", proxyConfig.String(), len(DefaultProxyPool))
	DefaultProxyPool <- proxyConfig
	Logger.Printf("add proxy to default pool success %v,now default proxy pool size %d", proxyConfig.String(), len(DefaultProxyPool))
}

func AddProxyConfigToAvailableProxyPool(proxyConfig *TProxyConfig) {
	if len(AvailableProxyPool) == MaxPoolSize {
		Logger.Printf("add proxy to full available pool %v", proxyConfig.String())
		tmpProxyConf := <-AvailableProxyPool
		Logger.Printf("discard proxy from available pool %v", tmpProxyConf.String())
	}

	Logger.Printf("will proxy to available pool %v,now available proxy pool size %d", proxyConfig.String(), len(AvailableProxyPool))
	AvailableProxyPool <- proxyConfig
	Logger.Printf("add proxy to available pool success %v,now available proxy pool size %d", proxyConfig.String(), len(AvailableProxyPool))
}

func GetOneProxyConfigFromProxyPool() *TProxyConfig {
	select {
	case proxyConfig := <-AvailableProxyPool:
		Logger.Printf("get proxy from available pool success %v,now available proxy pool size %d", proxyConfig.String(), len(AvailableProxyPool))
		return proxyConfig
	default:
	}

	if len(DefaultProxyPool) == 0 {
		for _, v := range GConfig.ProxyList {
			if v.Type != TProxyIsNone {
				AddProxyConfigToDefaultProxyPool(&v)
			}
		}
	}

	select {
	case proxyConfig := <-DefaultProxyPool:
		Logger.Printf("get proxy from default pool success %v,now default proxy pool size %d", proxyConfig.String(), len(DefaultProxyPool))
		return proxyConfig
	default:
		Logger.Print("get proxy from default pool fail")
		return &TProxyConfig{TProxyIsNone, ""}
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
		AddProxyConfigToDefaultProxyPool(&TProxyConfig{
			TProxyIsSocks,
			fmt.Sprintf("%s:%s", proxyItem.Address, proxyItem.Port),
		})
	case "http":
		AddProxyConfigToDefaultProxyPool(&TProxyConfig{
			TProxyIsHttp,
			fmt.Sprintf("%s:%s", proxyItem.Address, proxyItem.Port),
		})
	}
}
