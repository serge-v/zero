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

	mux := &http.ServeMux{}
	mux.HandleFunc("/email", server.EmailHandler)
	go func() {
		if err := http.ListenAndServe("127.0.0.1:8000", mux); err != nil {
			log.Fatal("service endpoint error", err)
		}
	}()

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
