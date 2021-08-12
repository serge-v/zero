package main

import (
	"embed"
	"flag"
	"log"
	"net/http"

	"github.com/serge-v/zero"
)

//go:embed *.html *.css
var content embed.FS

var deploy = flag.Bool("deploy", false, "")

func main() {
	flag.Parse()

	if *deploy {
		if err := zero.Deploy(8097); err != nil {
			log.Fatal(err)
		}
		return
	}

	http.Handle("/", http.FileServer(http.FS(content)))

	if err := http.ListenAndServe("127.0.0.1:8097", nil); err != nil {
		log.Fatal(err)
	}
}
