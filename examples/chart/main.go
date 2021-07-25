package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/kelvins/sunrisesunset"
	"github.com/serge-v/zero"
	"github.com/wcharczuk/go-chart/v2"
	"github.com/wcharczuk/go-chart/v2/drawing"

	_ "embed"
	_ "time/tzdata"
)

var deploy = flag.Bool("deploy", false, "")
var standalone = flag.Bool("standalone", false, "")

func main() {
	flag.Parse()

	if *standalone {
		if err := saveChart("1.png"); err != nil {
			log.Fatal(err)
		}
		return
	}

	if *deploy {
		if err := zero.Deploy(8096); err != nil {
			log.Fatal(err)
		}
		return
	}

	http.HandleFunc("/", handleRoot)
	http.HandleFunc("/chart/chart/chart.png", handleChart)

	if err := http.ListenAndServe("127.0.0.1:8096", nil); err != nil {
		log.Fatal(err)
	}
}

//go:embed main.html
var mainPage string

func handleRoot(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, mainPage)
}

func handleChart(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "image/png")
	generateChart(w)
}

func saveChart(fname string) error {
	f, err := os.Create(fname)
	if err != nil {
		return fmt.Errorf("save: %w", err)
	}
	defer f.Close()
	if err := generateChart(f); err != nil {
		return fmt.Errorf("generate: %w", err)
	}
	return nil
}

func generateChart(w io.Writer) error {
	resp, err := http.Get("http://localhost:8095/moisture.log")
	if err != nil {
		return fmt.Errorf("generate: %w", err)
	}
	defer resp.Body.Close()

	buf, err := ioutil.ReadAll(resp.Body)
	lines := strings.Split(string(buf), "\n")

	var xvalues []time.Time
	var yvalues []float64

	for _, line := range lines {
		cc := strings.Split(line, " ")
		if len(cc) < 8 {
			continue
		}
		timestr := cc[0] + " " + cc[1] + " " + cc[2]
		ts, err := time.Parse("2006-01-02 15:04:05 MST", timestr)
		if err != nil {
			return fmt.Errorf(": %w", err)
		}

		var moisture int

		for _, c := range cc {
			kv := strings.Split(c, ":")
			if len(kv) != 2 {
				continue
			}
			switch kv[0] {
			case "moisture":
				moisture, _ = strconv.Atoi(kv[1])
			}
		}

		xvalues = append(xvalues, ts)
		yvalues = append(yvalues, float64(moisture))
	}

	var ygrid []chart.GridLine

	for i := 100; i < 1000; i += 50 {
		ygrid = append(ygrid, chart.GridLine{Value: float64(i)})
	}

	graph := chart.Chart{
		Series: []chart.Series{
			chart.TimeSeries{
				XValues: xvalues,
				YValues: yvalues,
			},
		},
		YAxis: chart.YAxis{
			GridMajorStyle: chart.Style{
				StrokeColor: chart.ColorAlternateGray,
				StrokeWidth: 0.5,
			},
			GridLines: ygrid,
		},
		XAxis: chart.XAxis{
			ValueFormatter: chart.TimeHourValueFormatter,
			GridMajorStyle: chart.Style{
				StrokeColor: chart.ColorAlternateGray,
				StrokeWidth: 1.0,
			},
			GridLines: dayLines(),
			Ticks:     ticks(),
		},
	}
	if err := graph.Render(chart.PNG, w); err != nil {
		return fmt.Errorf("render: %w", err)
	}
	return nil
}

func ticks() []chart.Tick {
	var list []chart.Tick
	y, m, d := time.Now().Date()
	start := time.Date(y, m, d, 0, 0, 0, 0, time.UTC)

	for i := 0; i < 3; i++ {
		ts := start.Add(time.Hour * 24 * time.Duration(-i))
		line := chart.Tick{
			Value: chart.TimeToFloat64(ts),
			Label: ts.Format("1/02"),
		}
		list = append(list, line)

		ts = ts.Add(time.Hour * 6)
		line = chart.Tick{
			Value: chart.TimeToFloat64(ts),
			Label: ts.Format("15"),
		}
		list = append(list, line)

		ts = ts.Add(time.Hour * 6)
		line = chart.Tick{
			Value: chart.TimeToFloat64(ts),
			Label: ts.Format("15"),
		}
		list = append(list, line)

		ts = ts.Add(time.Hour * 6)
		line = chart.Tick{
			Value: chart.TimeToFloat64(ts),
			Label: ts.Format("15"),
		}
		list = append(list, line)
	}
	return list
}

func dayLines() []chart.GridLine {
	var list []chart.GridLine
	y, m, d := time.Now().Date()
	start := time.Date(y, m, d, 12, 0, 0, 0, time.UTC)

	for i := 0; i < 5; i++ {
		ts := start.Add(time.Hour * 24 * time.Duration(-i))
		line := chart.GridLine{
			Value: chart.TimeToFloat64(ts),
			Style: chart.Style{StrokeColor: drawing.ColorRed},
		}
		list = append(list, line)
	}

	start = time.Date(y, m, d, 0, 0, 0, 0, time.UTC)

	for i := 0; i < 5; i++ {
		ts := start.Add(time.Hour * 24 * time.Duration(-i))
		line := chart.GridLine{
			Value: chart.TimeToFloat64(ts),
			Style: chart.Style{StrokeColor: drawing.ColorBlue},
		}
		list = append(list, line)

		ts = start.Add(time.Hour*24*time.Duration(-i) + time.Hour*6)
		line = chart.GridLine{
			Value: chart.TimeToFloat64(ts),
			Style: chart.Style{StrokeColor: drawing.ColorBlue, StrokeWidth: 0.5},
		}
		list = append(list, line)

		ts = start.Add(time.Hour*24*time.Duration(-i) + time.Hour*18)
		line = chart.GridLine{
			Value: chart.TimeToFloat64(ts),
			Style: chart.Style{StrokeColor: drawing.ColorBlue, StrokeWidth: 0.5},
		}
		list = append(list, line)

		p := sunrisesunset.Parameters{
			Latitude:  41,
			Longitude: -74,
			UtcOffset: -4.0,
			Date:      ts,
		}

		sunrise, sunset, err := p.GetSunriseSunset()
		if err != nil {
			log.Fatal(err)
		}

		line = chart.GridLine{
			Value: chart.TimeToFloat64(sunrise),
			Style: chart.Style{StrokeColor: drawing.ColorFromHex("f5f242")},
		}
		list = append(list, line)

		line = chart.GridLine{
			Value: chart.TimeToFloat64(sunset),
			Style: chart.Style{StrokeColor: drawing.ColorFromHex("f5a742")},
		}
		list = append(list, line)

	}

	return list
}
