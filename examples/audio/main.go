package main

import (
	"embed"
	"flag"
	"html/template"
	"io"
	"io/fs"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	_ "embed"
	_ "time/tzdata"
)

type fit []fs.FileInfo

func (fi fit) Len() int {
	return len(fi)
}

func (fi fit) Swap(i, j int) {
	fi[i], fi[j] = fi[j], fi[i]
}

var rex = regexp.MustCompile("[0-9]+")

func (fi fit) Less(i, j int) bool {
	s1 := rex.FindString(fi[i].Name())
	s2 := rex.FindString(fi[j].Name())
	n1, _ := strconv.Atoi(s1)
	n2, _ := strconv.Atoi(s2)
	return n1 < n2
}

func getFiles(subdir string) ([]fs.FileInfo, error) {
	files, err := ioutil.ReadDir(filepath.Join(dir, subdir))
	if err != nil {
		return nil, err
	}

	sort.Sort(fit(files))
	return files, nil
}

//go:embed *.html *.js favicon.ico
var resourses embed.FS
var templates *template.Template

func reloadTemplates(w http.ResponseWriter) {
	var err error
	templates, err = template.ParseFiles("player.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func handlePlayer(w http.ResponseWriter, r *http.Request) {
	if *debug {
		reloadTemplates(w)
	}

	var err error

	d := struct {
		Dirs []string
	}{}

	d.Dirs, err = getDirs()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := templates.ExecuteTemplate(w, "player", d); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func getDirs() ([]string, error) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var list []string
	for _, f := range files {
		if f.IsDir() {
			list = append(list, f.Name())
		}
	}

	sort.Strings(list)
	return list, nil
}

func handleFileList(w http.ResponseWriter, r *http.Request) {
	subdir := r.URL.Query().Get("dir")
	files, err := getFiles(subdir)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var fnames []string
	for _, f := range files {
		if filepath.Ext(f.Name()) != ".mp3" && filepath.Ext(f.Name()) != ".webm" {
			continue
		}
		fnames = append(fnames, `	"`+url.PathEscape(filepath.Join(subdir, f.Name()))+`"`)
	}
	io.WriteString(w, "[\n"+strings.Join(fnames, ",\n")+"\n]\n")
}

var debug = flag.Bool("d", false, "debug")
var dir = "/audiofiles/"

func main() {
	flag.Parse()

	var err error

	templates, err = template.ParseFS(resourses, "player.html")
	if err != nil {
		log.Fatal(err)
	}

	if *debug {
		dir = "." + dir
	}

	res := http.FileServer(http.FS(resourses))
	http.Handle("/favicon.ico", res)
	http.Handle("/player.js", res)
	http.HandleFunc("/", handlePlayer)
	http.HandleFunc("/files", handleFileList)
	http.Handle("/audio/", http.StripPrefix("/audio/", http.FileServer(http.Dir(dir))))

	log.Println("starting")
	if err := http.ListenAndServe("127.0.0.1:8101", nil); err != nil {
		log.Fatal(err)
	}
}
