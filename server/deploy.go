package server

import (
	"encoding/json"
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
	"strings"
)

const appsDir = "/apps/"

func HandleDeployRequest(w http.ResponseWriter, r *http.Request) {
	if err := handleDeploy(w, r); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func HandleLogRequest(w http.ResponseWriter, r *http.Request) {
	if err := handleLog(w, r); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func handleDeploy(w http.ResponseWriter, r *http.Request) error {
	token := os.Getenv("TOKEN")
	rtoken := r.URL.Query().Get("token")
	if rtoken == "" || token != rtoken {
		return fmt.Errorf("invalid token")
	}

	port, err := getPort(r)
	if err != nil {
		return err
	}

	appname, err := saveTempApp(r, appsDir)
	if err != nil {
		return fmt.Errorf("appname: %w", err)
	}

	pidfile := appsDir + appname + ".pid"
	if err := stopApp(pidfile); err != nil {
		log.Println(err)
		os.Remove(pidfile)
	}

	if err := os.Rename(appsDir+appname+".tmp", appsDir+appname); err != nil {
		return fmt.Errorf("cannot rename tmp app: %w", err)
	}

	if err := startApp(appname, port); err != nil {
		return fmt.Errorf("start app: %w", err)
	}

	return nil
}

func handleLog(w http.ResponseWriter, r *http.Request) error {
	token := os.Getenv("TOKEN")
	rtoken := r.URL.Query().Get("token")
	if rtoken == "" || token != rtoken {
		return fmt.Errorf("invalid token")
	}
	appname := r.URL.Query().Get("appname")
	if appname == "" {
		return fmt.Errorf("no appname parameter")
	}
	appname = strings.ReplaceAll(appname, ".", "/")
	logname := "/tmp/" + appname + ".log"
	buf, _ := ioutil.ReadFile(logname)
	w.Write(buf)

	return nil
}

func startApp(appname string, port int) error {
	pidfile := appsDir + appname + ".pid"

	logname := "/tmp/" + appname + ".log"
	logf, err := os.Create(logname)
	if err != nil {
		log.Println(err)
		logf = os.Stderr
	}

	cmd := exec.Command("./" + appname)
	cmd.Dir = appsDir
	//	cmd.Env = []string{}
	cmd.Stderr = logf

	err = cmd.Start()
	if err != nil {
		return fmt.Errorf("start app %s: %w", appname, err)
	}

	pid := fmt.Sprintf("%d", cmd.Process.Pid)
	if err := ioutil.WriteFile(pidfile, []byte(pid), 0644); err != nil {
		return fmt.Errorf("write pid file for %s: %w", appname, err)
	}

	log.Println(appname, pid, "started")

	go cleanOnExit(cmd, pidfile, logf)

	appLock.Lock()
	defer appLock.Unlock()

	handler := apps[appname]
	if handler == nil {
		apps[appname] = createProxy(appname, port)
		log.Println("proxy created for app", appname, "port", port)
	}
	ports[appname] = port

	buf, err := json.Marshal(ports)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}

	if err := ioutil.WriteFile(appsDir+"ports.json", buf, 0o644); err != nil {
		return fmt.Errorf(": %w", err)
	}

	return nil
}

func StartApps() {
	buf, err := ioutil.ReadFile(appsDir + "ports.json")
	if err != nil {
		log.Println(err)
		return
	}

	if err := json.Unmarshal(buf, &ports); err != nil {
		log.Println(err)
		return
	}

	for appname, port := range ports {
		if err := startApp(appname, port); err != nil {
			log.Println(err)
		}
	}
}

func saveTempApp(r *http.Request, dir string) (string, error) {
	if err := r.ParseForm(); err != nil {
		return "", fmt.Errorf("parse form: %w", err)
	}

	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("mkdir: %w", err)
	}

	appname := r.URL.Query().Get("appname")
	appname = strings.ReplaceAll(appname, ".", "/")
	if appname == "" {
		return "", fmt.Errorf("no appname parameter")
	}

	fname := filepath.Join(dir, appname+".tmp")
	f, err := os.Create(fname)
	if err != nil {
		return "", fmt.Errorf("create: %w", err)
	}
	defer f.Close()

	n, err := io.Copy(f, r.Body)
	if err != nil {
		return "", fmt.Errorf("save app %s: %w", appname, err)
	}
	f.Close()

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

func cleanOnExit(cmd *exec.Cmd, pidfile string, logf io.Closer) {
	err := cmd.Wait()
	if err != nil {
		log.Println(err)
	}
	if logf != nil && logf != os.Stderr {
		logf.Close()
	}
	if err := os.Remove(pidfile); err != nil {
		log.Println(cmd.Path, pidfile, err)
	} else {
		log.Println("pid removed for", cmd.Path)
	}
}

func createProxy(appname string, port int) http.Handler {
	addr := fmt.Sprintf("http://127.0.0.1:%d/", port)
	u, err := url.Parse(addr)
	if err != nil {
		log.Fatal(err)
	}
	return httputil.NewSingleHostReverseProxy(u)
}

func getPort(r *http.Request) (int, error) {
	var err error
	var port int

	s := r.URL.Query().Get("port")
	if s == "" {
		return 0, fmt.Errorf("invalid port")
	}

	port, err = strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("parse port: %w", err)
	}
	return port, nil
}
