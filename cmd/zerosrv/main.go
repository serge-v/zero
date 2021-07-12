package main

import (
	"log"
	"net/http"

	"github.com/serge-v/zero/server"
)

func main() {
	http.HandleFunc("/", server.HandleDeployRequest)
	log.Println("starting")
	if err := http.ListenAndServe("127.0.0.1:8088", nil); err != nil {
		log.Fatal(err)
	}
}
