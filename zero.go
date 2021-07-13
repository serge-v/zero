package zero

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

func addFiles(values url.Values, patterns []string) error {
	for _, pattern := range patterns {
		files, err := filepath.Glob(pattern)
		if err != nil {
			return fmt.Errorf("glob for %s: %w", pattern, err)
		}
		for _, fname := range files {
			buf, err := ioutil.ReadFile(fname)
			if err != nil {
				return fmt.Errorf("read file %s: %w", fname, err)
			}
			values.Add(fname, string(buf))
		}
	}
	return nil
}

var durl = "https://zero.voilokov.com/"

//var durl = "http://127.0.0.1:8099/"

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

	buf, err := ioutil.ReadFile("token~.txt")
	if err != nil {
		return fmt.Errorf("read token: %w", err)
	}
	token := string(buf)

	params := fmt.Sprintf("deploy?appname=%s&token=%s&port=%d", appname, token, port)
	resp, err := http.Post(durl+params, "application/octet-stream", f)
	if err != nil {
		return fmt.Errorf("post to %s: %w", durl, err)
	}
	defer resp.Body.Close()

	buf, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read all: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("status: %s, resp:\n%s\n", resp.Status, string(buf))
	}

	return nil
}

func buildApp() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("cwd: %w", err)
	}

	appname := filepath.Base(dir)
	outdir, err := ioutil.TempDir(dir, "zero-*")
	if err != nil {
		return "", fmt.Errorf("temp dir: %w", err)
	}

	outname := filepath.Join(outdir, appname)

	cmd := exec.Command("go", "build", "-ldflags", "-X main.compileDate="+time.Now().Format("2006-01-02T15:04:05"), "-o", outname)
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "GOOS=linux", "GOARCH=amd64")
	log.Println("building:", dir, cmd.Args)

	buf, err := cmd.CombinedOutput()
	log.Println(string(buf))
	if err != nil {
		return string(buf), fmt.Errorf("run compiler: %w", err)
	}
	log.Println("done:", outname)
	return outname, nil
}
