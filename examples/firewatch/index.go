package main

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"firewatch/control44"
)

const indexPageText = `<!DOCTYPE html>
<html>
<head>
	<title>firewatch</title>
</head>
<body>
firewatch<br>
<a href="test">test</a>
</body>
</html>
`

func indexPage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, indexPageText)
}

var lock sync.Mutex

func testPage(w http.ResponseWriter, r *http.Request) {
	lock.Lock()
	defer lock.Unlock()

	text, _, err := control44.GetIncidents()
	if err != nil {
		text = "error:" + err.Error()
	} else if text == "" {
		text = "text is empty"
	}

	sendFirewatchEmail(text)
	time.Sleep(time.Second * 20)
}
