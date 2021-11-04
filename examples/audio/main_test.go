package main

import (
	"regexp"
	"testing"
)

func TestRegexp(t *testing.T) {
	r := regexp.MustCompile("[0-9]+")
	s := r.FindString("M111-aa.txt")
	t.Log(s)
}
