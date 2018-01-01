package main

import (
	"fmt"
	"net"
	"net/http"
	"net/http/fcgi"
)

type FastCGIServer struct{}

func (s FastCGIServer) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	fmt.Printf("%s\n", req.RemoteAddr)
	w.Write([]byte("Nice.\n"))
}

func main() {
	fmt.Println("Starting server...")
	l, _ := net.Listen("tcp", "127.0.0.1:9003")
	b := new(FastCGIServer)
	fcgi.Serve(l, b)
}
