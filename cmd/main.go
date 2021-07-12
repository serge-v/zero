package main

import (
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/", handleRequest)
	if err := http.ListenAndServe("127.0.0.1:8088", nil); err != nil {
		log.Fatal(err)
	}
}

func handleRequest(w http.ResponseWriter, r *http.Request) {

}
