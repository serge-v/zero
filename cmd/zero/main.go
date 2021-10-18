package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/serge-v/zero"
)

var deploy = flag.Int("deploy", 0, "port")
var showLog = flag.Bool("log", false, "show log")

func main() {
	flag.Parse()

	if *deploy != 0 {
		if err := zero.Deploy(*deploy); err != nil {
			log.Fatal(err)
		}
		if *showLog {
			time.Sleep(time.Second * 3)
		}
	}

	if *showLog {
		if text, err := zero.Log(); err != nil {
			log.Fatal(err)
		} else {
			fmt.Println(text)
		}
	}
}
