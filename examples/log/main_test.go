package main

import "testing"

func TestReadLastLines(t *testing.T) {
	lines, err := readLastLines("main.go")
	if err != nil {
		t.Fatal(err)
	}
	println("=== ln", len(lines))

	for _, ln := range lines {
		t.Log(ln)
	}
}
