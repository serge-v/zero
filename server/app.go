package server

import (
	"fmt"
	"log"
	"net/http"
	"sort"
	"strings"
	"sync"
)

var (
	appLock sync.Mutex
	apps    = make(map[string]http.Handler)
	ports   = make(map[string]int)
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
			fmt.Fprintf(w, `<a href="/%s/">%s</a><br>`+"\n", s, s)
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
	log.Println("appname", appname)
	appLock.Lock()
	handler := apps[appname]
	appLock.Unlock()

	if handler == nil {
		log.Println("not found", r.URL.Path)
		http.NotFound(w, r)
		return nil
	}

	handler = http.StripPrefix("/"+appname+"/", handler)
	handler.ServeHTTP(w, r)
	return nil
}
