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
	"net/url"
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

	client = Clients.AddNewClient(clientUUID, conn)

	defer func() {
		Clients.RemoveClient(clientUUID)
		Logger.Printf("client %s[%s] disconnected", conn.RemoteAddr(), clientUUID)
	}()

	qstr := r.URL.RawQuery
	if len(qstr) == 0 {
		qstr = Config.QueryString
	} else {
		qstr = fmt.Sprintf("%s&%s", qstr, Config.QueryString)
	}

	Logger.Printf("client %s[%s] final query[%s]", conn.RemoteAddr(), clientUUID, qstr)

	env := make(map[string]string)
	env["SCRIPT_FILENAME"] = Config.ScriptFileName
	env["QUERY_STRING"] = qstr

	for _, item := range Config.HeaderParams {
		env[item.Key] = item.Value
	}

	remoteInfo := strings.Split(conn.RemoteAddr().String(), ":")
	env["REMOTE_ADDR"] = remoteInfo[0]
	env["REMOTE_PORT"] = remoteInfo[1]

	env["PROXY_UUID"] = clientUUID

	body := bytes.NewReader(nil)

	for {
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			Logger.Printf("client %s[%s] read err message failed %s", conn.RemoteAddr(), clientUUID, err)
			break
		}

		if messageType != websocket.TextMessage {
			Logger.Printf("client %s[%s] read err message type", conn.RemoteAddr(),clientUUID)
			break
		}

		body.Reset(p)

		Logger.Printf("client %s[%s] request body [%s][%d]", conn.RemoteAddr(), clientUUID, string(p), body.Len())

		fcgi, err := fcgiclient.Dial("tcp", Config.FcgiServerAddress)
		if err != nil {
			Logger.Print(err)
			break
		}

		resp, err := fcgi.Post(env, "application/octet-stream", body, body.Len())
		if err != nil {
			Logger.Printf("client %s[%s] read fcgi response failed %s", conn.RemoteAddr(), clientUUID, err)
			fcgi.Close()
			break
		}

		content, err := ioutil.ReadAll(resp.Body)
		fcgi.Close()

		if err != nil {
			Logger.Printf("client %s[%s] read fcgi response failed %s", conn.RemoteAddr(), clientUUID, err)
			break
		}

		err = client.PushMessage(content)
		if err != nil {
			Logger.Printf("client %s[%s] response failed %s", conn.RemoteAddr(), clientUUID, err)
			break
		}
	}
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
