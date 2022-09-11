package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"text/template"
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
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	if *standalone {
		if err := saveChart("sensor1", "1.png"); err != nil {
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
	http.HandleFunc("/chart.png", handleChart)
	http.HandleFunc("/chart.svg", handleSVGChart)
	http.HandleFunc("/data.csv", handleCsv)
	http.HandleFunc("/events.csv", handleEvents)

	log.Println("starting on http://127.0.0.1:8096")

	if err := http.ListenAndServe("127.0.0.1:8096", nil); err != nil {
		log.Fatal(err)
	}
}

//go:embed main.html
var mainPage string

//go:embed chart.svg
var chartTemplate string

func handleRoot(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, mainPage)
}

func handleChart(w http.ResponseWriter, r *http.Request) {
	fname := r.URL.Query().Get("f")
	w.Header().Set("Content-Type", "image/png")
	generateChart(w, fname)
}

func handleSVGChart(w http.ResponseWriter, r *http.Request) {
	fname := r.URL.Query().Get("f")
	log.Println("chart svg")
	w.Header().Set("Content-Type", "image/svg+xml")
	if err := generateSVGChart(w, fname); err != nil {
		log.Println(err)
	}
}

func handleEvents(w http.ResponseWriter, r *http.Request) {
	now := time.Now().Truncate(time.Hour * 24).Add(time.Hour * 4)
	fmt.Fprint(w, "time,event\n")
	for i := -3; i < 1; i++ {
		ts := now.Add(time.Hour * 24 * time.Duration(i))
		fmt.Fprint(w, ts.Format(time.RFC3339), ",midnight", "\n")
		ts = now.Add(time.Hour*24*time.Duration(i) + time.Hour*12)
		fmt.Fprint(w, ts.Format(time.RFC3339), ",noon", "\n")
	}
}

func handleCsv(w http.ResponseWriter, r *http.Request) {
	fname := r.URL.Query().Get("f")
	lines, err := fetchSensorLog(fname)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	records, err := parseLogLines(lines)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Access-Control-Allow-Origin", "*")
	fmt.Fprint(w, "time,moisture,battery,hall\n")
	for _, r := range records {
		fmt.Fprint(w, r.ts.Add(time.Hour*4).Format(time.RFC3339), ",", r.moisture, ",", fmt.Sprintf("%.2f", r.battery), ",", r.hall, "\n")
	}
}

func saveChart(fname, outfname string) error {
	f, err := os.Create(outfname)
	if err != nil {
		return fmt.Errorf("save: %w", err)
	}
	defer f.Close()
	if err := generateChart(f, fname); err != nil {
		return fmt.Errorf("generate: %w", err)
	}
	return nil
}

func generateChart(w io.Writer, fname string) error {
	lines, err := fetchSensorLog(fname)
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

	st := chart.Style{DotWidth: 1, DotColor: chart.ColorBlue, StrokeWidth: chart.Disabled}
	lowHumidity := chart.Style{StrokeWidth: 3, StrokeColor: chart.ColorRed}

	for i := 300; i < 800; i += 50 {
		ygrid = append(ygrid, chart.GridLine{Value: float64(i)})
		ygrid = append(ygrid, chart.GridLine{Value: float64(i + 25), Style: st})
	}
	ygrid = append(ygrid, chart.GridLine{Value: 525, Style: lowHumidity})

	var yticks []chart.Tick
	for i := 300; i < 800; i += 50 {
		//		yticks = append(yticks, chart.Tick{Value: float64(i), Label: fmt.Sprintf("%d", i)})
	}

	yticks = append(yticks, chart.Tick{Value: 625, Label: "high"})
	yticks = append(yticks, chart.Tick{Value: 575, Label: "norm"})
	yticks = append(yticks, chart.Tick{Value: 525, Label: "low"})

	graph := chart.Chart{
		DPI: 144,
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

func generateSVGChart(w io.Writer, fname string) error {
	lines, err := fetchSensorLog(fname)
	if err != nil {
		return fmt.Errorf("fetch: %w", err)
	}

	/*
		buf, err := ioutil.ReadFile("sensor.log")
		if err != nil {
			return fmt.Errorf("read sensor log: %w", err)
		}
	*/
	//	lines = strings.Split(string(buf), "\n")

	records, err := parseLogLines(lines)
	if err != nil {
		return fmt.Errorf("parse: %w", err)
	}

	type line struct {
		X1, Y1, X2, Y2 int
		Class          string
	}

	type data struct {
		Lines     []line
		DayLabels []string
		Polyline  string
	}

	d := data{}
	year, month, day := time.Now().Add(time.Hour * 24).Date()
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		return err
	}

	start := time.Date(year, month, day, 0, 0, 0, 0, loc)

	for i := 0; i <= 5; i++ {
		ts := start.Add(time.Duration(-i) * time.Hour * 24).Format("1/2")
		d.DayLabels = append(d.DayLabels, ts)
	}

	if len(records) > 110 {
		records = records[len(records)-110:]
	}

	var prev line

	for i, r := range records {
		var ln line

		if r.moisture < 410 {
			r.moisture = 410
		}

		ln.X1 = int(start.Sub(r.ts).Minutes()) - 240
		ln.X2 = ln.X1 + 60
		ln.Y1 = r.moisture
		ln.Y2 = r.moisture
		ln.Class = "mhor"

		println("=== ", start.String(), r.ts.UTC().String(), ln.X1)

		if i > 0 {
			ln2 := line{X1: prev.X1, X2: prev.X1, Y1: prev.Y1, Y2: ln.Y1, Class: "mver"}
			d.Lines = append(d.Lines, ln2)
		}

		d.Lines = append(d.Lines, ln)
		prev = ln
	}

	var t *template.Template

	_, err = os.Stat("chart.svg")
	if err != nil {
		t, err = template.New("").Parse(chartTemplate)
	} else {
		t, err = template.ParseFiles("chart.svg")
	}
	if err != nil {
		return fmt.Errorf("parse template: %w", err)
	}

	if err := t.Execute(w, d); err != nil {
		return fmt.Errorf("execute: %w", err)
	}

	return nil
}
