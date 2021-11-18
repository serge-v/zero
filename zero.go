package zero

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

var durl = "https://zero.voilokov.com/"

// Deploy installs and starts golang app from the current directory to the zero server.
func Deploy(port int) error {
	fname, err := buildApp()
	if err != nil {
		return fmt.Errorf("build: %w", err)
	}
	appname := filepath.Base(fname)

	f, err := os.Open(fname)
	if err != nil {
		return fmt.Errorf("open: %w", err)
	}
	defer f.Close()

	token, err := getToken()
	if err != nil {
		return fmt.Errorf("read token: %w", err)
	}

	params := fmt.Sprintf("deploy?appname=%s&token=%s&port=%d", appname, token, port)
	resp, err := http.Post(durl+params, "application/octet-stream", f)
	if err != nil {
		return fmt.Errorf("post to %s: %w", durl, err)
	}
	defer resp.Body.Close()

	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read all: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("status: %s, resp:\n%s\n", resp.Status, string(buf))
	}

	return nil
}

// Log gets an app log from the zero server.
func Log() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("cwd: %w", err)
	}

	appname := filepath.Base(dir)
	token, err := getToken()
	if err != nil {
		return "", fmt.Errorf("read token: %w", err)
	}

	params := fmt.Sprintf("log?appname=%s&token=%s", appname, token)
	resp, err := http.Get(durl + params)
	if err != nil {
		return "", fmt.Errorf("post to %s: %w", durl, err)
	}
	defer resp.Body.Close()

	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read all: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("status: %s, resp:\n%s\n", resp.Status, string(buf))
	}

	return string(buf), nil
}

func getToken() (string, error) {
	buf, err := ioutil.ReadFile("token~.txt")
	if err != nil {
		return "", fmt.Errorf("read token: %w", err)
	}
	return string(buf), nil
}

func buildApp() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("cwd: %w", err)
	}

	appname := filepath.Base(dir)
	outdir, err := ioutil.TempDir("", "zero-*")
	if err != nil {
		return "", fmt.Errorf("temp dir: %w", err)
	}

	outname := filepath.Join(outdir, appname)

	cmd := exec.Command("go", "build", "-ldflags", "-X main.compileDate="+time.Now().Format("2006-01-02T15:04:05-MST"), "-o", outname)
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "GOOS=linux", "GOARCH=amd64")
	log.Println("building:", dir, cmd.Args)

	buf, err := cmd.CombinedOutput()
	log.Println(string(buf))
	if err != nil {
		return string(buf), fmt.Errorf("run compiler: %w", err)
	}
	log.Println("built:", outname)
	return outname, nil
}
