package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/serge-v/zero"
)

var deploy = flag.Bool("deploy", false, "")

func handleRegister(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	email := r.PostFormValue("email")
	password := r.PostFormValue("password")
	if email == "" {
		http.Error(w, "email is empty", http.StatusBadRequest)
		return
	}
	if password == "" {
		http.Error(w, "password is empty", http.StatusBadRequest)
		return
	}

	u, err := db.createUser(email, password)
	if err != nil {
		log.Println(err)
	}

	link := "http://localhost:8098/confirm?rand=" + u.RandomHash

	fmt.Println("registration email sent to " + email)
}

func handleHomePage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "home.html")
}

func handleCss(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "main.css")
}

func handleDemo(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "demo.html")
}

var db userDB

func main() {
	flag.Parse()

	if *deploy {
		if err := zero.Deploy(8097); err != nil {
			log.Fatal(err)
		}
		return
	}

	db = loadDB("/tmp/demo.json")

	http.HandleFunc("/", handleHomePage)
	http.HandleFunc("/main.css", handleCss)
	http.HandleFunc("/demo", handleDemo)
	http.HandleFunc("/register", handleRegister)

	if err := http.ListenAndServe("127.0.0.1:8098", nil); err != nil {
		log.Fatal(err)
	}
}
