package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/serge-v/zero"
	"github.com/wcharczuk/go-chart/v2"

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
	lines, err := fetchSensorLog()
	if err != nil {
		return fmt.Errorf("fetch: %w", err)
	}

	records, err := parseLogLines(lines)
	if err != nil {
		return fmt.Errorf("parse: %w", err)
	}

	var xvalues []time.Time
	var yvalues []float64

	for _, r := range records {
		xvalues = append(xvalues, r.ts)
		yvalues = append(yvalues, float64(r.moisture))
	}

	var ygrid []chart.GridLine
	for i := 100; i < 1000; i += 50 {
		ygrid = append(ygrid, chart.GridLine{Value: float64(i)})
	}

	var yticks []chart.Tick
	for i := 100; i < 1000; i += 50 {
		yticks = append(yticks, chart.Tick{Value: float64(i), Label: fmt.Sprintf("%d", i)})
	}

	graph := chart.Chart{
		DPI: 96,
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
			Ticks:     yticks,
			Range: &chart.ContinuousRange{
				Min: 200,
				Max: 800,
			},
		},
		XAxis: chart.XAxis{
			ValueFormatter: chart.TimeHourValueFormatter,
			GridMajorStyle: chart.Style{
				StrokeColor: chart.ColorAlternateGray,
				StrokeWidth: 1.0,
			},
			GridLines: dayEventsLines(),
			Ticks:     xTicks(),
		},
	}
	if err := graph.Render(chart.PNG, w); err != nil {
		return fmt.Errorf("render: %w", err)
	}
	return nil
}
