package main

import (
	"log"
	"strings"
	"time"

	"firewatch/control44"

	_ "time/tzdata"
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
		log.Println("tick", h, m)

		if m == 7 {
			text, latest, err := control44.GetIncidents()
			if err != nil {
				log.Println(err)
			} else if text != "" && latest.After(latestSent) {
				sendFirewatchEmail(text)
				latestSent = latest
				time.Sleep(time.Second * 10)
			} else {
				log.Println("firewatch is empty or already sent. latest:", latest)
			}
		}
	}
}

func sendFirewatchEmail(body string) {
	msg := "From: Firewatch\n"
	msg += "To: firewatch users\n"
	msg += "Subject: Firewatch\n\n"
	msg += body

	users := strings.Fields(firewatchUsers)

	if err := smtpSend("Firewatch", users, []byte(msg)); err != nil {
		log.Println("sendmail:", err)
	}
}
