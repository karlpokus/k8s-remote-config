package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"

	"github.com/karlpokus/k8s-remote-config/conf"
)

var host = flag.String("h", "localhost", "HTTP host")
var port = flag.String("p", "7000", "HTTP port")
var confPath = flag.String("conf", "/etc/config", "Path to config file")

func main() {
	flag.Parse()
	b, err := os.ReadFile(*confPath)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("config: %+v", conf.Marshal(b))
	err = startServer(net.JoinHostPort(*host, *port))
	if err != nil {
		log.Printf("server start err: %v", err)
	}
}

func startServer(addr string) error {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "hello from server!")
	})
	log.Printf("starting HTTP server on %s", addr)
	return http.ListenAndServe(addr, nil)
}
