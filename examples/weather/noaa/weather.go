// test program to parse dwml xml format from NOAA API

package noaa

import (
	"bytes"
	"fmt"
	"log"
)

func PrintWeather(zip string) {
	var plain, html bytes.Buffer
	err := Forecast(&plain, &html, zip)
	if err != nil {
		log.Println(err)
	}
	fmt.Println(plain.String())
}
