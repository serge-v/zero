package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/serge-v/zero"

	_ "embed"
	_ "time/tzdata"
)

var list []string
var lastEmail time.Time

//go:embed log_users~.txt
var logUsers string

func appendLog(fname, s string) error {
	nyc, err := time.LoadLocation("America/New_York")
	if err != nil {
		return err
	}

	logfname := filepath.Clean("/data/" + fname + ".log")
	msg := time.Now().In(nyc).Format("2006-01-02 15:04:05 MST ") + s
	list = append(list, msg)
	f, err := os.OpenFile(logfname, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	n, err := fmt.Fprintln(f, msg)
	log.Println("written", msg, n, "to", logfname)
	if err != nil {
		return err
	}
	return nil
}

func handleAppendLog(w http.ResponseWriter, r *http.Request) {
	msg := r.URL.Query().Get("m")
	fname := r.URL.Query().Get("f")
	if fname == "" {
		fname = "sensor1"
	}

	if msg != "" {
		if err := appendLog(fname, msg); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if fname == "" {
			if err := appendLog("moisture", msg); err != nil { // TODO: remove after all sensor are upgraded
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
	}
}

var deploy = flag.Bool("deploy", false, "")

func main() {
	flag.Parse()

	if *deploy {
		if err := zero.Deploy(8095); err != nil {
			log.Fatal(err)
		}
		return
	}

	http.HandleFunc("/", handleAppendLog)
	http.HandleFunc("/moisture.log", handleLog)
	http.HandleFunc("/log", handleSensorLog)

	log.Println("starting on http://127.0.0.1:8095")
	if err := http.ListenAndServe("127.0.0.1:8095", nil); err != nil {
		log.Fatal(err)
	}
}
func readLastLines(fname string) ([]string, error) {
	f, err := os.Open(fname)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	f.Seek(-50000, os.SEEK_END)

	buf := make([]byte, 50000)
	n, err := f.Read(buf)
	if err != nil {
		return nil, err
	}

	log.Println("read", fname, n)

	lines := strings.Split(string(buf[:n]), "\n")
	if len(lines) > 0 {
		lines = lines[1:]
	}
	if len(lines) > 500 {
		lines = lines[len(lines)-500:]
	}

	return lines, nil
}

func handleSensorLog(w http.ResponseWriter, r *http.Request) {
	fname := r.URL.Query().Get("f")
	logfname := "/data/" + fname + ".log"
	logfname = filepath.Clean(logfname)
	lines, err := readLastLines(logfname)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	for _, ln := range lines {
		fmt.Fprintln(w, ln)
	}
}

func handleLog(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "/data/moisture.log")
}
