package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/serge-v/zero/server"
)

var debug = flag.Bool("debug", false, "run in debug mode")

func main() {
	flag.Parse()
	http.HandleFunc("/deploy", server.HandleDeployRequest)
	http.HandleFunc("/", server.HandleAppRequest)
	addr := ":80"
	if *debug {
		addr = "127.0.0.1:8099"
	}
	log.Println("starting on", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal(err)
	}
}
