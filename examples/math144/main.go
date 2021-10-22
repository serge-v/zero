// Handler math144 prints match exercises pages.
package main

import (
	"bytes"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

const page = `
<html>
<head>
<style>
    body {
        margin: 0;
        padding: 0;
        font: 14pt "Menlo";
        line-height: 1em;
    }
    pre {
        font: 14pt "Menlo";
        line-height: 1em;
    }
    * {
        box-sizing: border-box;
        -moz-box-sizing: border-box;
    }
    .page {
        width: 8.5in;
        min-height: 11in;
        padding: 0.8in;
        margin: 1in auto;
        border: 1px #D3D3D3 solid;
        border-radius: 5px;
        background: white;
    }
    .subpage {
        padding: 0in;
        border: 0px gray solid;
        height: 9in;
    }
    
    @page {
        size: Letter;
        margin: 0;
    }
    @media print {
        .page {
            margin: 0;
            border: initial;
            border-radius: initial;
            width: initial;
            min-height: initial;
            box-shadow: initial;
            background: initial;
            page-break-after: always;
        }
    }
</style>
</head>
<body>
<div class="book">
{pages}
</div>
</body>
</html>
`

func generateMultiplications(max int) string {
	var text string
	var b bytes.Buffer
	for x := 0; x < 7; x++ {
		for y := 0; y < 8; y++ {
			fmt.Fprintf(&b, " %2d     ", mathRand.Intn(max-3)+3)
		}
		fmt.Fprintln(&b)
		for y := 0; y < 8; y++ {
			fmt.Fprintf(&b, "x%2d     ", mathRand.Intn(max-3)+3)
		}
		fmt.Fprintln(&b)
		for y := 0; y < 8; y++ {
			fmt.Fprintf(&b, "---     ")
		}
		for y := 0; y < 5; y++ {
			fmt.Fprintln(&b)
		}
	}
	text += `<div class="page"><div class="subpage"><pre>` + b.String() + `</pre></div></div>`
	return text
}

func generateDivisions(max int) string {
	var text string
	var b bytes.Buffer
	for x := 0; x < 8; x++ {
		for y := 0; y < 6; y++ {
			fmt.Fprintf(&b, "   ----   ")
		}
		fmt.Fprintln(&b)
		for y := 0; y < 6; y++ {
			c := mathRand.Intn(max-2) + 2
			d := mathRand.Intn(max-2) + 2
			fmt.Fprintf(&b, " %2d) %2d   ", c, c*d)
		}
		for y := 0; y < 5; y++ {
			fmt.Fprintln(&b)
		}
	}
	text += `<div class="page"><div class="subpage"><pre>` + b.String() + `</pre></div></div>` + "\n"
	return text
}

var mathRand *rand.Rand

func init() {
	seed := time.Now().Truncate(time.Hour * 24 * 7).Unix() // new seed every week
	mathRand = rand.New(rand.NewSource(seed))
}

func HandleMath144(w http.ResponseWriter, r *http.Request) {
	var text string
	for page := 0; page < 10; page++ {
		text += generateMultiplications(13)
		text += generateDivisions(13)
	}
	out := strings.Replace(page, "{pages}", text, 1)
	fmt.Fprintln(w, out)
}

func main() {
	http.HandleFunc("/", HandleMath144)
	log.Fatal(http.ListenAndServe("127.0.0.1:8092", nil))
}
