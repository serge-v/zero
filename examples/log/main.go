package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/serge-v/zero"

	_ "time/tzdata"
)

var list []string
var lastEmail time.Time

func handler(w http.ResponseWriter, r *http.Request) {
	nyc, err := time.LoadLocation("America/New_York")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	msg := r.URL.Query().Get("m")
	if msg != "" {
		msg = time.Now().In(nyc).Format("2006-01-02 15:04:05 MST ") + msg
		list = append(list, msg)
		f, err := os.OpenFile("/data/moisture.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer f.Close()
		n, err := fmt.Fprintln(f, msg)
		log.Println("written", msg, n)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		f.Close()
	}

	if len(list) > 0 && time.Since(lastEmail) > time.Hour {
		text := list[len(list)-1]
		text, err := url.QueryUnescape(text)
		if err != nil {
			text = err.Error()
		}
		if err := zero.Email("soilsensor", []string{"voilokov@gmail.com"}, "moisture", text); err != nil {
			log.Println("email error:", err)
		}
		lastEmail = time.Now()
	}

	for i := len(list) - 1; i >= 0; i-- {
		fmt.Fprintln(w, list[i])
	}
}

var deploy = flag.Bool("deploy", false, "")

func main() {
	flag.Parse()

	if *deploy {
		if err := zero.Deploy(8095); err != nil {
			log.Fatal(err)
		}
		return
	}

	http.HandleFunc("/", handler)
	http.HandleFunc("/moisture.log", handleLog)

	if err := http.ListenAndServe("127.0.0.1:8095", nil); err != nil {
		log.Fatal(err)
	}
}

func handleLog(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "/data/moisture.log")
}
