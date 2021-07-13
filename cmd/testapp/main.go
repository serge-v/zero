package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/serge-v/zero"
)

var compileDate string

var deploy = flag.Bool("deploy", false, "deploy the app to the zero runner")

func main() {
	flag.Parse()
	if *deploy {
		if err := zero.Deploy(8091); err != nil {
			log.Fatal(err)
		}
		return
	}

	log.Println("=== test app on 127.0.0.1:8091")
	http.HandleFunc("/", handler)
	err := http.ListenAndServe("127.0.0.1:8091", nil)
	if err != nil {
		log.Fatal(err)
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "<h1>Test</h1>")
	fmt.Fprintln(w, "<p>compiled", compileDate)
	fmt.Fprintln(w, "<p>now", time.Now())
}
