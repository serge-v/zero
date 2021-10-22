package main

import (
	"log"
	"net/http"

	_ "embed"
	_ "time/tzdata"
)

//go:embed weather_users~.txt
var weatherUsers string

func main() {
	http.HandleFunc("/", indexPage)
	http.HandleFunc("/wx", handleWeather)

	go startJobs()

	addr := "127.0.0.1:8090"
	log.Println("starting on", addr)
	err := http.ListenAndServe(addr, nil)
	if err != nil {
		log.Fatal(err)
	}
}
