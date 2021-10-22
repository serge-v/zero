package control44

import (
	"bytes"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"

	_ "time/tzdata"
)

var (
	reTR  = regexp.MustCompile("(?sU)<tr.*</tr>")
	reTD  = regexp.MustCompile("(?sU)<td.*</td>")
	reTag = regexp.MustCompile("(?sU)<.*>")
)

func parseActiveIncidents(buf []byte) ([]incident, error) {
	if bytes.Contains(buf, []byte("NoIncidents")) {
		return nil, nil
	}

	matches := reTR.FindAllSubmatch(buf, -1)
	var list []incident
	if len(matches) < 2 {
		return nil, fmt.Errorf("invalid number of rows %d", len(matches))
	}
	for _, mm := range matches[2:] {
		for _, m := range mm {
			inc := parseActiveRow(m)
			list = append(list, inc)
		}
	}
	return list, nil
}

func parsePastIncidents(buf []byte) ([]incident, error) {
	matches := reTR.FindAllSubmatch(buf, -1)
	var list []incident
	if len(matches) < 2 {
		return nil, fmt.Errorf("invalid number of rows: %d", len(matches))
	}
	for _, mm := range matches[2:] {
		for _, m := range mm {
			inc := parseRow(m)
			list = append(list, inc)
		}
	}
	return list, nil
}

var nyc = func() *time.Location {
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		log.Fatal(err)
	}
	return loc
}()

func guessTime(s string) time.Time {
	ts, err := time.Parse("01/02  15:04", s)
	if err != nil {
		log.Fatal(err)
	}
	ts = time.Date(time.Now().Year(), ts.Month(), ts.Day(), ts.Hour(), ts.Minute(), ts.Second(), 0, nyc)
	if ts.After(time.Now()) {
		// since timestamp has no year guess this is a last year
		ts = time.Date(time.Now().Year()-1, ts.Month(), ts.Day(), ts.Hour(), ts.Minute(), ts.Second(), 0, nyc)
	}
	return ts
}

func parseRow(buf []byte) incident {
	cells := reTD.FindAllSubmatch(buf, -1)
	if len(cells) != 7 {
		log.Fatal("invalid line", string(buf), len(cells))
	}

	var inc incident

	for i, mm := range cells {
		var m []byte
		var s string
		if len(mm) == 1 {
			m = reTag.ReplaceAll(mm[0], []byte{})
		}
		s = strings.TrimSpace(string(m))
		switch i {
		case 0:
			inc.Dispatched = guessTime(s)
		case 1:
			inc.Closed = guessTime(s)
		case 2:
			inc.FireDepartment, _ = strconv.Atoi(s)
		case 3:
			inc.Type = s
		case 4:
			inc.Number, _ = strconv.Atoi(s)
		case 5:
			inc.Address = s
		case 6:
			if s != "&nbsp;" {
				inc.CommonName = s
			}
		}
	}

	return inc
}

func parseActiveRow(buf []byte) incident {
	cells := reTD.FindAllSubmatch(buf, -1)
	if len(cells) != 8 {
		log.Fatal("invalid line", string(buf), len(cells))
	}

	var inc incident

	for i, mm := range cells {
		var m []byte
		var s string
		if len(mm) == 1 {
			m = reTag.ReplaceAll(mm[0], []byte{})
		}
		s = strings.TrimSpace(string(m))
		switch i {
		case 0:
			inc.Dispatched = guessTime(s)
		case 1:
			inc.FireDepartment, _ = strconv.Atoi(s)
		case 2:
			inc.Type = s
		case 3:
			inc.Number, _ = strconv.Atoi(s)
		case 4:
			inc.Address = s
		case 6:
			if s != "&nbsp;" {
				inc.CommonName = s
			}
		case 7:
			if s != "&nbsp;" {
				inc.CallStatus = s
			}
		}
	}

	return inc
}
