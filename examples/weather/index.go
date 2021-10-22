package main

import (
	"fmt"
	"net/http"
)

const indexPageText = `<!DOCTYPE html>
<html>
<head>
	<title>weather</title>
</head>
<body>
	<a href="wx?zip=11230">wx-11230</a><br>
	<a href="wx?zip=00804">wx-00804</a><br>
	<a href="wx?zip=12309">wx-12309</a><br>
	<a href="wx?zip=10974">wx-10974</a>
	<a href="wx?zip=10974&send=1">send</a><br>
</body>
</html>
`

func indexPage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, indexPageText)
}
