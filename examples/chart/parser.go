package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type record struct {
	ts       time.Time
	moisture int
	battery  float64
	hall     int
}

func fetchSensorLog() ([]string, error) {
	resp, err := http.Get("https://zero.voilokov.com/log/moisture.log")
	if err != nil {
		return nil, fmt.Errorf("generate: %w", err)
	}
	defer resp.Body.Close()

	buf, err := ioutil.ReadAll(resp.Body)
	lines := strings.Split(string(buf), "\n")
	return lines, nil
}

func parseLogLines(lines []string) ([]record, error) {
	var list []record

	for _, line := range lines {
		cc := strings.Split(line, " ")
		if len(cc) < 8 {
			continue
		}

		var err error
		var r record

		timestr := cc[0] + " " + cc[1] + " " + cc[2]
		r.ts, err = time.Parse("2006-01-02 15:04:05 MST", timestr)
		if err != nil {
			return nil, fmt.Errorf(": %w", err)
		}

		for _, c := range cc {
			kv := strings.Split(c, ":")
			if len(kv) != 2 {
				continue
			}
			switch kv[0] {
			case "moisture":
				r.moisture, _ = strconv.Atoi(kv[1])
			case "hall":
				r.hall, _ = strconv.Atoi(kv[1])
			case "batv":
				r.battery, _ = strconv.ParseFloat(strings.TrimSuffix(kv[1], "V"), 64)
			}
		}

		list = append(list, r)
	}
	return list, nil
}
