// Package noaa defines noaa.gov dwml API structures.
package noaa

import (
	"encoding/xml"
	"errors"
	"log"
	"time"
)

var (
	// ErrNOAAError is a server error.
	ErrNOAAError = errors.New("NOAA server error")
	// ErrBadXML is a xml error.
	ErrBadXML = errors.New("bad dwml xml")
)

// Dwml is a top level element.
type Dwml struct {
	XMLName xml.Name `xml:"dwml"`
	Header  Header   `xml:"head"`
	Data    Data     `xml:"data"`
}

// Header is a Dwml.Header.
type Header struct {
	XMLName xml.Name `xml:"head"`
	Product Product  `xml:"product"`
}

// Data is a Dwml.Data.
type Data struct {
	XMLName     xml.Name     `xml:"data"`
	TimeLayouts []TimeLayout `xml:"time-layout"`
	Parameters  Parameters   `xml:"parameters"`
}

func (d *Data) getTimeLayout(key string) []time.Time {
	var layout *TimeLayout
	for _, v := range d.TimeLayouts {
		if v.Key == key {
			layout = &v
			break
		}
	}

	if layout == nil {
		log.Fatalf("empty layout: %s", key)
	}

	times := make([]time.Time, len(layout.StartTime))
	for idx, v := range layout.StartTime {
		startTs, err := time.Parse("2006-01-02T15:04:05-07:00", v)
		if err != nil {
			log.Fatalln("parse time", err)
		}
		times[idx] = startTs
	}

	return times
}

// Product is a Header.Product.
type Product struct {
	XMLName xml.Name `xml:"product"`
	Src     string   `xml:"srsName,attr"`
	Name    string   `xml:"concise-name,attr"`
	Mode    string   `xml:"operational-mode,attr"`
}

// TimeLayout is a Data.TimeLayout.
type TimeLayout struct {
	XMLName       xml.Name `xml:"time-layout"`
	Coordinate    string   `xml:"time-coordinate,attr"`
	Summarization string   `xml:"summarization,attr"`
	Key           string   `xml:"layout-key"`
	StartTime     []string `xml:"start-valid-time"`
	EndTime       []string `xml:"end-valid-time"`
}

// Parameters is a Data.Parameters.
type Parameters struct {
	XMLName       xml.Name   `xml:"parameters"`
	Temperature   []Valueset `xml:"temperature"`
	WindSpeed     []Valueset `xml:"wind-speed"`
	Direction     []Valueset `xml:"direction"`
	CloudAmount   []Valueset `xml:"cloud-amount"`
	Precipitation []Valueset `xml:"precipitation"`
	Humidity      []Valueset `xml:"humidity"`
	Weather       Weather    `xml:"weather"`
	Hazards       Hazards    `xml:"hazards"`
}

// Visibility is a Data.Parameters.
type Visibility struct {
	XMLName xml.Name
	Units   string `xml:"units,attr"`
}

// ConditionValue is a Weather.WeatherConditions.ConditionValue.
type ConditionValue struct {
	Coverage    string `xml:"coverage,attr"`
	Intensity   string `xml:"intensity,attr"`
	WeatherType string `xml:"weather-type,attr"`
	Qualifier   string `xml:"qualifier,attr"`
}

// WeatherConditions is a Weather.WeatherConditions.
type WeatherConditions struct {
	Value []ConditionValue `xml:"value"`
}

// Weather is a Data.Parameters.Weather.
type Weather struct {
	TimeLayout string              `xml:"time-layout,attr"`
	Conditions []WeatherConditions `xml:"weather-conditions"`
}

// Valueset ...
type Valueset struct {
	Type       string   `xml:"type,attr"`
	Units      string   `xml:"units,attr"`
	TimeLayout string   `xml:"time-layout,attr"`
	Name       string   `xml:"name"`
	Values     []string `xml:"value"`
}

// Hazards is a Data.Parameters.Hazards
type Hazards struct {
	TimeLayout       string           `xml:"time-layout,attr"`
	Name             string           `xml:"name"`
	HazardConditions HazardConditions `xml:"hazard-conditions"`
}

// HazardConditions is a Data.Parameters.Hazards.HazardConditions
type HazardConditions struct {
	Hazard []Hazard `xml:"hazard"`
}

// Hazard is a Data.Parameters.Hazards.HazardConditions.Hazard
type Hazard struct {
	HazardTextURL       string `xml:"hazardTextURL"`
	HazardIcon          string `xml:"hazardIcon"`
	HazardCode          string `xml:"hazardCode,attr"`
	Phenomena           string `xml:"phenomena,attr"`
	Significance        string `xml:"significance,attr"`
	HazardType          string `xml:"hazardType,attr"`
	EventTrackingNumber int    `xml:"eventTrackingNumber,attr"`
	Headline            string `xml:"headline,attr"`
}
