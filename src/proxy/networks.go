package proxy

import (
	"net/http"
	"net/url"
	"time"
	"net"
	"math/rand"
	"golang.org/x/net/proxy"
)

func GetOneProxyConfig() TProxyConfig {
	if len(Config.ProxyList) == 0 {
		return TProxyConfig{TProxyIsNone, ""}
	}

	proxyIndex := rand.Intn(len(Config.ProxyList))
	return Config.ProxyList[proxyIndex]
}

func MakeHttpClient() (*http.Client, error) {
	var httpClient *http.Client

	proxyConfig := GetOneProxyConfig()

	switch proxyConfig.Type {
	case TProxyIsHttp:
		proxyUrl, err := url.Parse(proxyConfig.Address)
		if err != nil {
			return nil, err
		}

		httpClient = &http.Client{
			Transport: &http.Transport{
				Proxy:           http.ProxyURL(proxyUrl),
				IdleConnTimeout: 10 * time.Second,
			},
		}
		break
	case TProxyIsSocks:
		dialer, err := proxy.SOCKS5(
			"tcp",
			proxyConfig.Address,
			nil,
			&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			},
		)

		if err != nil {
			return nil, err
		}

		httpClient = &http.Client{
			Transport: &http.Transport{
				Dial:     dialer.Dial,
				IdleConnTimeout: 10 * time.Second,
			},
		}
		break
	case TProxyIsNone:
		httpClient = &http.Client{
			Transport: &http.Transport{
				IdleConnTimeout: 10 * time.Second,
			},
		}
		break
	}

	return httpClient, nil
}