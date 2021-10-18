package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/serge-v/zero"
)

var deploy = flag.Int("deploy", 0, "port")

func main() {
	flag.Parse()

	if *deploy != 0 {
		if err := zero.Deploy(*deploy); err != nil {
			log.Fatal(err)
		}
		time.Sleep(time.Second * 5)
		if text, err := zero.Log(); err != nil {
			log.Fatal(err)
		} else {
			fmt.Println(text)
		}
		return
	}

	flag.Usage()
}
