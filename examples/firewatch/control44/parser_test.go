package control44

import (
	"fmt"
	"testing"
)

func TestParser(t *testing.T) {
	t.Log("log")
	text, ts, err := GetIncidents()
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(ts, text)
}
