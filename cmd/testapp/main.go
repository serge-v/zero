package main

import (
	"embed"
	"flag"
	"fmt"
	"log"
	"net/http"
	"text/template"

	_ "embed"

	"github.com/serge-v/zero"
)

var compileDate string

var deploy = flag.Bool("deploy", false, "deploy the app to the zero runner")

func main() {
	flag.Parse()
	if *deploy {
		if err := zero.Deploy(8091); err != nil {
			log.Fatal(err)
		}
		return
	}

	http.HandleFunc("/", handler)
	err := http.ListenAndServe("127.0.0.1:8091", nil)
	if err != nil {
		log.Fatal(err)
	}
}

//go:embed *.go go.mod
var files embed.FS

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "<h1>This is test application</h1>")
	fmt.Fprintln(w, "<p>compiledDate", compileDate)

	dirs, err := files.ReadDir(".")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	for _, de := range dirs {
		fmt.Fprintf(w, "<h3>%s</h3>\n", de.Name())
		buf, err := files.ReadFile(de.Name())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		fmt.Fprintf(w, "<pre>%s</pre>\n\n", template.HTMLEscapeString(string(buf)))
	}
}
