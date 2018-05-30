package proxy

import (
	"net/http"
	"net/url"
	"time"
	"net"
	"golang.org/x/net/proxy"
)

type THttpClient struct {
	httpClient *http.Client
	proxyIndex int
}

func MakeHttpClient(index int) (*THttpClient, error) {
	var httpClient *http.Client

	proxyConfig, proxyIndex := Config.GetOneProxyConfig(index)

	Logger.Printf("select proxy server %s", proxyConfig.String())

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
				Timeout:   10 * time.Second,
				KeepAlive: 10 * time.Second,
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

	return &THttpClient{httpClient, proxyIndex}, nil
}