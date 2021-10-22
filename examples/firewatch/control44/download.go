package control44

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

func download(from, to string) error {
	fi, err := os.Stat(to)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("stat file %s: %w", to, err)
	}

	if err == nil && time.Since(fi.ModTime()) < time.Second*600 {
		log.Println("file", to, "is recent. age:", time.Since(fi.ModTime()))
		return nil
	}

	f, err := os.Create(to)
	if err != nil {
		return fmt.Errorf("open file %s: %w", to, err)
	}

	defer f.Close()

	resp, err := http.Get(from)
	if err != nil {
		return fmt.Errorf("download %s: %w", from, err)
	}

	defer resp.Body.Close()

	n, err := io.Copy(f, resp.Body)
	if err != nil {
		return fmt.Errorf("copy from %s to %s: %w", from, to, err)
	}

	log.Println("downloaded", n, "bytes from", from, "into", to)

	return nil
}
