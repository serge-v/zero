package main

import (
	"log"
	"net/http"

	_ "embed"
	_ "time/tzdata"
)

//go:embed firewatch_users~.txt
var firewatchUsers string

func main() {
	http.HandleFunc("/", indexPage)
	http.HandleFunc("/test", testPage)

	addr := "127.0.0.1:8100"
	log.Println("starting on", addr)

	go startJobs()

	err := http.ListenAndServe(addr, nil)
	if err != nil {
		log.Fatal(err)
	}
}
