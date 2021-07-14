package server

import (
	"fmt"
	"net/http"
)

func EmailHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Example: curl -F from=... -F to=... -F subject=... -F body=... http://127.0.0.1:8000/email", http.StatusBadRequest)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	from := r.Form.Get("from")
	to := r.Form.Get("to")
	subject := r.Form.Get("subject")
	body := r.Form.Get("body")

	text := "From: " + from + "\n"
	text += "To: " + to + "\n"
	text += "Subject: " + subject + "\n\n"
	text += body + "\n"

	err := smtpSend("zerosrv", []string{"voilokov@gmail.com"}, []byte(text))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "sent email: %d", len(text))
}
