package proxy

import (
	"context"
	"golang.org/x/net/proxy"
	"net"
	"net/http"
	"net/url"
	"time"
)

type THttpClient struct {
	httpClient  *http.Client
	proxyConfig *TProxyConfig
}

func (obj *THttpClient) Success() {
	if obj.proxyConfig.Type == TProxyIsNone {
		return
	}

	AddProxyConfigToAvailableProxyPool(obj.proxyConfig)
}

func MakeHttpClient() (*THttpClient, error) {
	var httpClient *http.Client

	proxyConfig := GetOneProxyConfigFromProxyPool()

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
			proxy.Direct,
		)

		if err != nil {
			return nil, err
		}

		httpClient = &http.Client{
			Transport: &http.Transport{
				DialContext: func(_ context.Context, network, addr string) (net.Conn, error) {
					return dialer.Dial(network, addr)
				},
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

	return &THttpClient{httpClient, proxyConfig}, nil
}
