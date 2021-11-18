package main

import (
	"embed"
	"log"
	"net/http"
)

//go:embed *.html *.css
var content embed.FS

func main() {
	http.Handle("/", http.FileServer(http.FS(content)))
	if err := http.ListenAndServe("127.0.0.1:8097", nil); err != nil {
		log.Fatal(err)
	}
}
