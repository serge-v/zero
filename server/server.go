package server

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
)

func HandleDeployRequest(w http.ResponseWriter, r *http.Request) {
	if err := handleDeploy(w, r); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func handleDeploy(w http.ResponseWriter, r *http.Request) error {
	dir := "/tmp/builder/"

	if err := saveFiles(r, dir); err != nil {
		return fmt.Errorf("deploy: %w", err)
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

	log.Println("uploaded", appname, n, "bytes")

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
