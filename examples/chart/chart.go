package main

import (
	"log"
	"time"

	"github.com/kelvins/sunrisesunset"
	"github.com/wcharczuk/go-chart/v2"
	"github.com/wcharczuk/go-chart/v2/drawing"
)

// creates 6 hours grid with NYC sunrise and sunset lines.
func dayEventsLines() []chart.GridLine {
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

func xTicks() []chart.Tick {
	var list []chart.Tick

	start := time.Now().Truncate(time.Hour * 6).Add(-time.Hour * 6)

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

		ts = ts.Add(time.Hour*6 + time.Hour*3)
		line = chart.Tick{
			Value: chart.TimeToFloat64(ts),
		}
		list = append(list, line)

		ts = ts.Add(time.Hour*6 - time.Hour*3)
		line = chart.Tick{
			Value: chart.TimeToFloat64(ts),
		}
		list = append(list, line)
	}
	return list
}
