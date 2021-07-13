package server

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"sync"
)

func HandleDeployRequest(w http.ResponseWriter, r *http.Request) {
	if err := handleDeploy(w, r); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

var (
	appLock sync.Mutex
	apps    = make(map[string]http.Handler)
)

func createProxy(appname string, port int) http.Handler {
	addr := fmt.Sprintf("http://127.0.0.1:%d/%s/", port, appname)
	u, err := url.Parse(addr)
	if err != nil {
		log.Fatal(err)
	}
	return httputil.NewSingleHostReverseProxy(u)
}

func handleDeploy(w http.ResponseWriter, r *http.Request) error {
	token := os.Getenv("TOKEN")
	rtoken := r.URL.Query().Get("token")
	if rtoken == "" || token != rtoken {
		return fmt.Errorf("invalid token")
	}

	dir := "/tmp/apps/"

	var err error
	var port int

	s := r.URL.Query().Get("port")
	if s == "" {
		return fmt.Errorf("invalid port")
	}

	port, err = strconv.Atoi(s)
	if err != nil {
		return fmt.Errorf("parse port: %w", err)
	}

	appname, err := saveApp(r, dir)
	if err != nil {
		return fmt.Errorf("appname: %w", err)
	}

	pidfile := dir + appname + ".pid"
	if err := stopApp(pidfile); err != nil {
		log.Println(err)
		os.Remove(pidfile)
	}

	cmd := exec.Command("./" + appname)
	cmd.Dir = dir
	cmd.Stderr = os.Stderr

	err = cmd.Start()
	if err != nil {
		return fmt.Errorf("start app %s: %w", appname, err)
	}

	pid := fmt.Sprintf("%d", cmd.Process.Pid)
	if err := ioutil.WriteFile(pidfile, []byte(pid), 0644); err != nil {
		return fmt.Errorf("write pid file for %s: %w", appname, err)
	}

	log.Println(appname, pid, "started")

	go cleanOnExit(cmd, pidfile)

	appLock.Lock()
	defer appLock.Unlock()

	handler := apps[appname]
	if handler == nil {
		apps[appname] = createProxy(appname, port)
		log.Println("proxy created for app", appname, "port", port)
	}

	return nil
}

func saveApp(r *http.Request, dir string) (string, error) {
	if err := r.ParseForm(); err != nil {
		return "", fmt.Errorf("parse form: %w", err)
	}

	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("mkdir: %w", err)
	}

	appname := r.URL.Query().Get("appname")
	if appname == "" {
		return "", fmt.Errorf("no appname parameter")
	}

	fname := filepath.Join(dir, appname)
	f, err := os.Create(fname)
	if err != nil {
		return "", fmt.Errorf("create: %w", err)
	}

	n, err := io.Copy(f, r.Body)
	if err != nil {
		return "", fmt.Errorf("save app %s: %w", appname, err)
	}

	log.Println("uploaded", fname, n, "bytes")

	if err := os.Chmod(fname, 0700); err != nil {
		return "", fmt.Errorf("chmod: %w", err)
	}

	return appname, nil
}

func saveFiles(r *http.Request, dir string) error {
	if err := r.ParseForm(); err != nil {
		return fmt.Errorf("parse form: %w", err)
	}

	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("mkdir: %w", err)
	}

	for fname, text := range r.PostForm {
		if err := ioutil.WriteFile(dir+fname, []byte(text[0]), 0644); err != nil {
			return fmt.Errorf("save file: %w", err)
		}
	}
	return nil
}

func stopApp(pidfile string) error {
	buf, _ := ioutil.ReadFile(pidfile)
	if len(buf) == 0 {
		return nil
	}
	pid, err := strconv.Atoi(string(buf))
	if err != nil {
		return fmt.Errorf("convert pid: %w", err)
	}

	pr, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("find process: %w", err)
	}

	if err := pr.Signal(os.Interrupt); err != nil {
		return fmt.Errorf("send signal: %w", err)
	}
	return nil
}

func cleanOnExit(cmd *exec.Cmd, pidfile string) {
	err := cmd.Wait()
	if err != nil {
		log.Println(err)
	}
	if err := os.Remove(pidfile); err != nil {
		log.Println(cmd.Path, pidfile, err)
	} else {
		log.Println("pid removed for", cmd.Path)
	}
}
