package server

import (
	"fmt"
	"log"
	"net/http"
	"sort"
	"strings"
)

func HandleAppRequest(w http.ResponseWriter, r *http.Request) {
	if err := handleApp(w, r); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func handleApp(w http.ResponseWriter, r *http.Request) error {
	path := r.URL.Path
	if path == "/" {
		var list []string
		for k := range apps {
			list = append(list, k)
		}
		sort.Strings(list)
		for _, s := range list {
			fmt.Fprintf(w, `<a href="/%s/">%s`+"\n", s, s)
		}
		return nil
	}

	log.Println("request", path)
	path = strings.TrimPrefix(path, "/")
	cc := strings.SplitN(path, "/", 2)
	if len(cc) != 2 {
		log.Println("not found", r.URL.Path)
		http.NotFound(w, r)
		return nil
	}
	appname := cc[0]
	handler := apps[appname]
	if handler == nil {
		log.Println("not found", r.URL.Path)
		http.NotFound(w, r)
		return nil
	}

	handler.ServeHTTP(w, r)
	return nil
}
