package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"weather/noaa"
)

var lastSent time.Time

func handleWeather(w http.ResponseWriter, r *http.Request) {
	zip := r.URL.Query().Get("zip")
	if zip == "" {
		http.Error(w, "invalid zip parameter", http.StatusBadRequest)
		return
	}

	send := r.URL.Query().Get("send")
	if send == "1" {
		if time.Since(lastSent) < time.Minute {
			http.Error(w, "too many requests", http.StatusServiceUnavailable)
			return
		}
		lastSent = time.Now()
		if err := sendWeather(zip); err != nil {
			log.Println(err)
			http.Error(w, "send error", http.StatusInternalServerError)
			return
		}
		fmt.Fprintln(w, "ok")
		return
	}

	var plain, html bytes.Buffer
	if err := noaa.Forecast(&plain, &html, zip); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	io.WriteString(w, html.String())
}
