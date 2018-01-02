package proxy

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		origin := strings.ToLower(r.Header.Get("Origin"))
		Logger.Printf("client %s Origin=%s request websocket server", r.RemoteAddr, origin)
		if InStringArray("*", Config.Origins) || InStringArray(origin, Config.Origins) {
			return true
		}

		return false
	},
}

func defaultHttpHandle(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "version: %s", VERSION)
}

func proxyHttpHandle(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		Logger.Print(err)
		return
	}

	defer conn.Close()

	rv, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		Logger.Print(err)
		return
	}

	clientUUID := rv.Get("uuid")
	if len(clientUUID) == 0 {
		clientUUID = uuid.New().String()
	}

	Logger.Printf("client %s[%s] connected with query[%s]", conn.RemoteAddr(), clientUUID, r.URL.RawQuery)

	client := Clients.GetClient(clientUUID)
	if client != nil {
		client.Close()
	}

	client = Clients.AddNewClient(clientUUID, conn, r)

	defer func() {
		Clients.RemoveClient(clientUUID)
		Logger.Printf("client %s[%s] disconnected", conn.RemoteAddr(), clientUUID)
	}()

	go client.PipeSendMessage()
	client.PipeReadMessage()
}

func NewWebSocket() (*http.Server, chan int) {
	http.HandleFunc("/", defaultHttpHandle)
	http.HandleFunc("/proxy", proxyHttpHandle)

	httpServer := &http.Server{
		Addr:           Config.HttpServerAddress,
		Handler:        http.DefaultServeMux,
		ReadTimeout:    30 * time.Second,
		WriteTimeout:   30 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	httpStop := make(chan int)

	go func() {
		Logger.Printf("http server will run at %s", Config.HttpServerAddress)
		err := httpServer.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			Logger.Print(err)
		}
		Logger.Printf("http server stop at %s", Config.HttpServerAddress)
		httpStop <- 1
	}()

	return httpServer, httpStop
}
