// Package control44 fetches Rockland fire alerts.
package control44

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"
)

const (
	pastURL   = "http://www.44-control.net/History.aspx"
	activeURL = "http://www.44-control.net/default.aspx"
)

type incident struct {
	Dispatched     time.Time
	Closed         time.Time
	FireDepartment int
	Type           string
	Number         int // Type number
	Address        string
	CommonName     string
	Active         bool
	CallStatus     string
}

func splitAddress(address string) (string, string) {
	ss := strings.Split(address, ",")
	var town, addr string
	for i, s := range ss {
		s = strings.Title(strings.ToLower(strings.TrimSpace(s)))
		if i == len(ss)-1 {
			town = s
			break
		}
		if s != "" {
			addr += s + " "
		}
	}
	return town, strings.TrimSpace(addr)
}

func GetIncidents() (string, time.Time, error) {
	dir := ".cache/control44/"
	active := dir + "active.html"
	past := dir + "past.html"

	os.MkdirAll(dir, 0777)

	if err := download(activeURL, active); err != nil {
		return "", time.Time{}, fmt.Errorf("download active %s: %w", pastURL, err)
	}

	if err := download(pastURL, past); err != nil {
		return "", time.Time{}, fmt.Errorf("download past %s: %w", pastURL, err)
	}

	var err error
	var list []incident
	var latest time.Time

	buf, err := ioutil.ReadFile(active)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("read file %s: %w", active, err)
	}

	var b bytes.Buffer

	if list, err = parseActiveIncidents(buf); err != nil {
		return "", time.Time{}, fmt.Errorf("parse active %s: %w", pastURL, err)
	}

	var count int

	fmt.Fprintf(&b, "Active:\n")
	for _, item := range list {
		if item.FireDepartment == 15 || strings.Contains(item.Address, "SLOATSBURG") {
			count++
			if latest.Before(item.Dispatched) {
				latest = item.Dispatched
			}
			fmt.Fprintf(&b, "%s %s, type: %s\n", item.Dispatched, item.Address, item.Type)
		}
	}

	if count == 0 {
		b.Reset()
	}
	count = 0

	buf, err = ioutil.ReadFile(past)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("read file %s: %w", past, err)
	}

	if list, err = parsePastIncidents(buf); err != nil {
		return "", time.Time{}, fmt.Errorf("parse past %s: %w", pastURL, err)
	}

	fmt.Fprintf(&b, "Past:\n")
	for _, item := range list {
		if item.FireDepartment == 15 || strings.Contains(item.Address, "SLOATSBURG") {
			if latest.Before(item.Dispatched) {
				latest = item.Dispatched
			}
			fmt.Fprintf(&b, "%s %s, type: %s\n", item.Dispatched, item.Address, item.Type)
			count++
		}
	}

	if count == 0 {
		b.Reset()
	}

	return b.String(), latest, nil
}
