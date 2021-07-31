package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os/exec"

	"github.com/serge-v/zero"
)

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, r.URL.Path)

	cmd := exec.Command("/bin/netstat", "-lp")
	cmd.Stdout = w
	if err := cmd.Run(); err != nil {
		fmt.Fprintln(w, err.Error())
	}

	cmd = exec.Command("/bin/ls", "-lp", "/data")
	cmd.Stdout = w
	if err := cmd.Run(); err != nil {
		fmt.Fprintln(w, err.Error())
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
