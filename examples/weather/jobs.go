package main

import (
	"bytes"
	"fmt"
	"log"
	"strings"
	"time"

	"weather/noaa"
)

var latestSent time.Time

var loc = func() *time.Location {
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		panic("cannot load location")
	}
	return loc
}()

func startJobs() {
	log.Println("starting jobs")
	ticker := time.NewTicker(time.Second * 60)

	for tick := range ticker.C {
		local := tick.In(loc)
		h, m := local.Hour(), local.Minute()
		if h == 6 && m == 5 {
			for i := 0; i < 5; i++ {
				log.Println("sending weather email")
				if err := sendWeather("10974"); err != nil {
					log.Println(err)
					time.Sleep(time.Second * 2)
					continue
				}
				break
			}
			time.Sleep(time.Second * 10)
		}
	}
}

func sendWeather(zip string) error {
	var plain, html bytes.Buffer
	err := noaa.Forecast(&plain, &html, zip)
	if err != nil {
		return fmt.Errorf("get forecast: %w", err)
	}

	from := "Weather sender"
	to := "Weather users"
	users := strings.Fields(weatherUsers)

	msg := "From: " + from + "\n"
	msg += "To: " + to + "\n"
	msg += "Subject: wx " + zip + "\n"
	text, err := formatMultipartMessage(&plain, &html, zip)
	if err != nil {
		return fmt.Errorf("format message: %w", err)
	}

	msg += text

	if err := smtpSend("Weather sender", users, []byte(msg)); err != nil {
		return fmt.Errorf("smtp send: %w", err)
	}

	return nil
}
