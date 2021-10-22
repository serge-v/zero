package noaa

import (
	"embed"
	"encoding/xml"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

// Parse decodes dwml forecast file.
func Parse(xmlFile *os.File) (*Dwml, error) {
	decoder := xml.NewDecoder(xmlFile)
	p := &Dwml{}
	if err := decoder.Decode(p); err != nil {
		return nil, fmt.Errorf("cannot decode dwml file: %w", err)
	}
	return p, nil

}

var debug = true

// TestFname is dwml xml file name for test.
var TestFname string

// NumRetries is max num retries.
var NumRetries = 3

//go:embed weather.html
var htmls embed.FS

func fileCached(fname string) bool {
	fi, err := os.Stat(fname)

	if err != nil || fi == nil {
		return false
	}

	if debug || (fi.ModTime().Unix()-time.Now().UTC().Unix() < 60*72) {
		return true
	}

	newname := fname + ".old"

	err = os.Rename(fname, newname)
	if err != nil {
		fmt.Printf("ERROR: %v\n", err)
	}

	return false
}

var cacheDir string

func init() {
	cacheDir = ".cache/weather/"
	os.MkdirAll(cacheDir, 0755)
}

func downloadForecast(zip string) (string, error) {
	now := time.Now()
	if TestFname != "" {
		return TestFname, nil
	}

	fname := fmt.Sprintf("%szip-%s-%s.xml", cacheDir, zip, now.Format("20060102-15"))
	if fileCached(fname) {
		return fname, nil
	}

	url := "http://www.weather.gov/forecasts/xml/SOAP_server/ndfdXMLclient.php?whichClient=NDFDgen&zipCodeList=%s&product=time-series&maxt=maxt&mint=mint&temp=temp&wspd=wspd&wdir=wdir&wx=wx&rh=rh&snow=snow&wwa=wwa&sky=sky&appt=appt&Submit=Submit&Unit=m"
	resp, err := http.Get(fmt.Sprintf(url, zip))
	if err != nil {
		return "", fmt.Errorf("cannot get url %s: %w", url, err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("cannot read body for url %s: %w", url, err)
	}
	err = ioutil.WriteFile(fname, body, 0666)
	if err != nil {
		return "", fmt.Errorf("cannot write file %s: %w", fname, err)
	}
	return fname, nil
}

// Forecast prints forecast for specified zip to the writer.
func Forecast(w io.Writer, whtml io.Writer, zip string) error {
	var p *Dwml
	var err error

	for i := 0; i < NumRetries; i++ {
		if i > 0 {
			log.Println("sleep 10 seconds")
			time.Sleep(time.Second * 10)
		}
		var fname string
		var f *os.File
		fname, err = downloadForecast(zip)
		if err != nil {
			err = fmt.Errorf("cannot download: %w", err)
			log.Println(err)
			continue
		}
		f, err = os.Open(fname)
		if err != nil {
			err = fmt.Errorf("cannot open file: %w", err)
			log.Println(err)
			continue
		}
		p, err = Parse(f)
		f.Close()
		if err != nil {
			err = fmt.Errorf("cannot parse: %w", err)
			log.Println(err)
			continue
		}
		err = nil
		break
	}

	if err != nil {
		s := err.Error()
		if strings.Contains(s, "expected element type <dwml> but have <error>") {
			return ErrNOAAError
		}
		if strings.Contains(s, "XML syntax error") {
			return ErrBadXML
		}
		return fmt.Errorf("forecast failed after retrying %d times: %w", NumRetries, err)
	}

	minStartTime := "9999"
	for _, v := range p.Data.TimeLayouts {
		for _, st := range v.StartTime {
			if strings.Compare(st, minStartTime) < 0 {
				minStartTime = st
			}
		}
	}
	minTs, err := time.Parse("2006-01-02T15:04:05-07:00", minStartTime)
	if err != nil {
		return fmt.Errorf("parse time: %w", err)
	}

	type temperature struct {
		Hourly   string
		Apparent string
		Max      string
		Min      string
	}

	type wind struct {
		Speed     int
		Direction int
	}

	type row struct {
		Time        time.Time
		Hour        string
		Date        string
		Midnight    bool
		Temperature temperature
		Wind        wind
		Clouds      int
		Hazards     string
		Snow        string
		Rain        string
		Weather     string
		HasData     bool
	}

	type units struct {
		Temperature string
		WindSpeed   string
		SnowAmount  string
	}

	const hours = 24 * 10
	rows := make([]row, hours)
	var uts units

	for i := 0; i < hours; i++ {
		r := row{Time: minTs.Add(time.Hour * time.Duration(i))}
		if r.Time.Hour() == 0 {
			r.Midnight = true
		}
		r.Hour = r.Time.Format("15")
		r.Date = r.Time.Format("2006-01-02 Mon")
		rows[i] = r
	}

	// fill temperature

	for _, t := range p.Data.Parameters.Temperature {
		uts.Temperature = t.Units
		times := p.Data.getTimeLayout(t.TimeLayout)
		if len(times) != len(t.Values) {
			return fmt.Errorf("invalid layout: %s", t.TimeLayout)
		}
		for idx, v := range t.Values {
			rowidx := int(times[idx].Sub(minTs).Hours())
			//			println(rowidx, startTs.String())
			switch t.Type {
			case "hourly":
				rows[rowidx].Temperature.Hourly = v
				rows[rowidx].HasData = true
			case "apparent":
				rows[rowidx].Temperature.Apparent = v
				rows[rowidx].HasData = true
			case "minimum":
				rows[rowidx].Temperature.Min = v
				rows[rowidx].HasData = true
			case "maximum":
				rows[rowidx].Temperature.Max = v
				rows[rowidx].HasData = true
			}
		}
	}

	// fill hazards

	if p.Data.Parameters.Hazards.TimeLayout != "" {
		times := p.Data.getTimeLayout(p.Data.Parameters.Hazards.TimeLayout)
		for idx, hz := range p.Data.Parameters.Hazards.HazardConditions.Hazard {
			rowidx := int(times[idx].Sub(minTs).Hours())
			text := hz.Phenomena + " " + hz.Significance
			rows[rowidx].Hazards = text
			rows[rowidx].HasData = true
		}
	}

	// fill wind-speed

	for _, vs := range p.Data.Parameters.WindSpeed {
		uts.WindSpeed = vs.Units
		times := p.Data.getTimeLayout(vs.TimeLayout)
		dir := p.Data.Parameters.Direction[0].Values
		for idx, v := range vs.Values {
			rowidx := int(times[idx].Sub(minTs).Hours())
			rows[rowidx].Wind.Speed, _ = strconv.Atoi(v)
			if len(dir) > idx {
				rows[rowidx].Wind.Direction, _ = strconv.Atoi(dir[idx])
			}
			rows[rowidx].HasData = true
		}
	}

	// fill clouds

	for _, vs := range p.Data.Parameters.CloudAmount {
		times := p.Data.getTimeLayout(vs.TimeLayout)
		for idx, v := range vs.Values {
			rowidx := int(times[idx].Sub(minTs).Hours())
			rows[rowidx].Clouds, _ = strconv.Atoi(v)
			rows[rowidx].HasData = true
		}
	}

	// fill precipitation

	for _, vs := range p.Data.Parameters.Precipitation {
		uts.SnowAmount = vs.Units
		times := p.Data.getTimeLayout(vs.TimeLayout)
		for idx, v := range vs.Values {
			rowidx := int(times[idx].Sub(minTs).Hours())
			if vs.Type == "snow" {
				rows[rowidx].Snow = v
			} else if vs.Type == "liquid" {
				rows[rowidx].Rain = v
			} else {
				rows[rowidx].Rain = v + "-" + vs.Type
			}
			rows[rowidx].HasData = true
		}
	}

	// fill weather
	for idx, cond := range p.Data.Parameters.Weather.Conditions {
		times := p.Data.getTimeLayout(p.Data.Parameters.Weather.TimeLayout)
		for _, v := range cond.Value {
			rowidx := int(times[idx].Sub(minTs).Hours())
			q := ""
			if v.Qualifier != "none" {
				q = v.Qualifier
			}
			intensity := v.Intensity
			if intensity == "heavy" {
				intensity = "HW"
			} else if intensity == "light" {
				intensity = "LT"
			} else if intensity == "very light" {
				intensity = "VLT"
			} else if intensity == "moderate" {
				intensity = "MOD"
			} else if intensity == "none" {
				intensity = ""
			}

			coverage := v.Coverage
			if coverage == "slight chance" {
				coverage = "20%"
			} else if coverage == "chance" {
				coverage = "40%"
			} else if coverage == "likely" {
				coverage = "60%"
			} else if coverage == "definitely" {
				coverage = "100%"
			} else if coverage == "isolated" {
				coverage = "ISO"
			}

			wxtype := v.WeatherType
			if wxtype == "rain" {
				wxtype = "R"
			} else if wxtype == "rain showers" {
				wxtype = "RSH"
			} else if wxtype == "snow" {
				wxtype = "SN"
			} else if wxtype == "snow showers" {
				wxtype = "SNSH"
			} else if wxtype == "thunderstorms" {
				wxtype = "TS"
			}
			rows[rowidx].Weather += ", " + coverage + " " + intensity + " " + wxtype + " " + q
			rows[rowidx].Weather = strings.Trim(rows[rowidx].Weather, " ,")
			rows[rowidx].HasData = true
		}
	}

	fmt.Fprintf(w, "hr |  tC | atC | minC| maxC| wnd-dir| cld | snow | rain | weather, hazards\n")
	for i := 0; i < hours; i++ {
		r := rows[i]
		if r.Time.Hour() == 0 {
			fmt.Fprintf(w, "---+-----+-----+-----+-----+--------+-----+------+------+- %s ----\n", r.Time.Format("2006-01-02 Mon"))
		}
		if !r.HasData {
			continue
		}
		t := r.Temperature
		wind := ""
		clouds := ""
		snow := ""
		if r.Wind.Speed > 0 {
			wind = fmt.Sprintf("%d-%03d", r.Wind.Speed, r.Wind.Direction)
		}
		if r.Clouds > 0 {
			clouds = fmt.Sprintf("%3d", r.Clouds)
		}
		if r.Snow == "0.00" {
			snow = ""
		}
		wx := r.Weather + ";" + r.Hazards
		wx = strings.Replace(strings.Trim(wx, "; "), "  ", " ", -1)
		fmt.Fprintf(w, "%s | %3s | %3s | %3s | %3s | %6s | %3s | %4s | %4s | %s\n",
			rows[i].Time.Format("15"), t.Hourly, t.Apparent, t.Min, t.Max,
			wind, clouds, snow, r.Rain,
			wx)
	}

	funcs := template.FuncMap{
		"nozero": func(n int) string {
			if n == 0 {
				return ""
			}
			return strconv.Itoa(n)
		},
		"nozeroprep": func(s string) string {
			if s == "0.00" {
				return ""
			}
			return s
		},
		"colortemp": func(s string) template.HTML {
			s = template.HTMLEscapeString(s)
			n, cerr := strconv.Atoi(s)
			if cerr != nil {
				return template.HTML(s)
			}
			if n >= 35 {
				return template.HTML(`<span style="background-color: coral">` + s + `</span>`)
			} else if n >= 25 {
				return template.HTML(`<span style="background-color: pink">` + s + `</span>`)
			} else if n < 5 {
				return template.HTML(`<span style="background-color: plum">` + s + `</span>`)
			} else if n < 15 {
				return template.HTML(`<span style="background-color: powderblue">` + s + `</span>`)
			}
			return template.HTML(s)
		},
		"windchar": func(n int) string {
			n = (n + 15) / 30
			if n == 12 {
				n = 0
			}
			return strconv.Itoa(n)
		},
	}

	var wxtmpl *template.Template

	wxtmpl, err = template.New("weather.html").Funcs(funcs).ParseFS(htmls, "weather.html")
	if err != nil {
		return fmt.Errorf("template: %w", err)
	}

	type data struct {
		Units   units
		Rows    []row
		Name    string
		Version string
		Date    string
	}

	hostname, _ := os.Hostname()
	d := data{
		Units:   uts,
		Rows:    rows,
		Name:    "test",
		Version: "1.0 " + hostname,
		Date:    time.Now().Format(time.RFC3339),
	}

	if err = wxtmpl.Execute(whtml, d); err != nil {
		return fmt.Errorf("parse template: %w", err)
	}

	return nil
}
