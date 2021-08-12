package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/serge-v/zero"
)

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, r.URL.Path)

	cmd := exec.Command("/bin/ls", "-lph", "/data")
	cmd.Stdout = w
	if err := cmd.Run(); err != nil {
		fmt.Fprintln(w, err.Error())
	}

	cmd = exec.Command("/bin/ls", "-lph", "/apps")
	cmd.Stdout = w
	if err := cmd.Run(); err != nil {
		fmt.Fprintln(w, err.Error())
	}

	cmd = exec.Command("/bin/ps", "-ef")
	cmd.Stdout = w
	if err := cmd.Run(); err != nil {
		fmt.Fprintln(w, err.Error())
	}

	cmd = exec.Command("/bin/cat", "/apps/ports.json")
	cmd.Stdout = w
	if err := cmd.Run(); err != nil {
		fmt.Fprintln(w, err.Error())
	}
	fmt.Fprintln(w, "")

	files, err := filepath.Glob("/apps/*")
	if err != nil {
		fmt.Fprintln(w, err.Error())
	}

	for _, fname := range files {
		if strings.HasSuffix(fname, ".pid") {
			continue
		}
		fmt.Fprintln(w, "app:", fname)
	}
}

var deploy = flag.Bool("deploy", false, "")

func main() {
	flag.Parse()

	if *deploy {
		if err := zero.Deploy(8094); err != nil {
			log.Fatal(err)
		}
		return
	}

	http.HandleFunc("/", handler)

	if err := http.ListenAndServe("127.0.0.1:8094", nil); err != nil {
		log.Fatal(err)
	}
}
