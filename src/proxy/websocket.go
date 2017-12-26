package proxy

import (
	"bytes"
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/tomasen/fcgi_client"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
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

	uuid := uuid.New().String()
	Clients.AddNewClient(uuid, conn)

	Logger.Printf("client %s[%s] connected", conn.RemoteAddr(), uuid)

	env := make(map[string]string)
	env["SCRIPT_FILENAME"] = Config.ScriptFileName
	env["QUERY_STRING"] = Config.QueryString

	for _, item := range Config.HeaderParams {
		env[item.Key] = item.Value
	}

	remoteInfo := strings.Split(conn.RemoteAddr().String(), ":")
	env["REMOTE_ADDR"] = remoteInfo[0]
	env["REMOTE_PORT"] = remoteInfo[1]

	env["PROXY_UUID"] = uuid

	body := bytes.NewReader(nil)

	for {
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			Logger.Print(err)
			return
		}

		if messageType != websocket.TextMessage {
			Logger.Printf("read err message type from %s", conn.RemoteAddr())
			return
		}

		body.Reset(p)

		Logger.Printf("request :%s  len=%d", string(p), body.Len())

		fcgi, err := fcgiclient.Dial("tcp", Config.PhpServerAddress)
		if err != nil {
			Logger.Print(err)
			return
		}

		resp, err := fcgi.Post(env, "application/octet-stream", body, body.Len())
		if err != nil {
			Logger.Printf("read php response failed %s", err)
			return
		}

		content, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			Logger.Printf("read php response failed %s", err)
			return
		}

		err = conn.WriteMessage(messageType, content)
		if err != nil {
			Logger.Printf("response failed %s", err)
			return
		}

		fcgi.Close()
	}

	Clients.RemoveClient(uuid)

	Logger.Print("client %s[%s] disconnected", conn.RemoteAddr(), uuid)
}

func WebSocket(ctx context.Context) *http.Server {
	http.HandleFunc("/", defaultHttpHandle)
	http.HandleFunc("/proxy", proxyHttpHandle)

	httpServer := &http.Server{
		Addr:           Config.HttpServerAddress,
		Handler:        http.DefaultServeMux,
		ReadTimeout:    30 * time.Second,
		WriteTimeout:   30 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	go func() {
		Logger.Printf("http server will run at %s", Config.HttpServerAddress)
		err := httpServer.ListenAndServe()
		if err != nil {
			Logger.Print(err)
		}
		Logger.Printf("http server will stop at %s", Config.HttpServerAddress)
	}()

	return httpServer
}
