package server

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

func buildApp(dir string) (string, error) {
	log.Println("building", dir)
	cmd := exec.Command("go", "build", "-ldflags", "-X main.compileDate="+time.Now().Format("2006-01-02T15:04:05"))
	cmd.Dir = dir
	buf, err := cmd.CombinedOutput()
	log.Println(string(buf))
	if err != nil {
		return string(buf), fmt.Errorf("run compiler: %w", err)
	}
	log.Println("done")
	return "", nil
}

func getAppName(dir string) (string, error) {
	buf, err := ioutil.ReadFile(dir + "go.mod")
	if err != nil {
		return "", fmt.Errorf("read go.mod: %w", err)
	}

	lines := strings.Split(string(buf), "\n")
	if len(lines) < 1 {
		return "", fmt.Errorf("invalid go.mod file")
	}

	s := strings.TrimPrefix(lines[0], "module ")
	cc := strings.Split(s, "/")
	if len(cc) == 0 {
		return "", fmt.Errorf("invalid go.mod file")
	}

	return cc[len(cc)-1], nil
}

func handleCompileAndDeploy(w http.ResponseWriter, r *http.Request) error {
	dir := "/tmp/builder/"

	if err := saveFiles(r, dir); err != nil {
		return fmt.Errorf("deploy: %w", err)
	}

	appname, err := getAppName(dir)
	if err != nil {
		return fmt.Errorf("appname: %w", err)
	}

	log.Println("appname:", appname)

	msgs, err := buildApp(dir)
	if err != nil {
		return fmt.Errorf("deploy: %w, messages: %s", err, msgs)
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

func HandleCompileAndDeployRequest(w http.ResponseWriter, r *http.Request) {
	if err := handleCompileAndDeploy(w, r); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
