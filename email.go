package zero

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
)

func Email(from string, to []string, subject, body string) error {
	values := url.Values{}
	values.Add("from", from)
	values.Add("to", strings.Join(to, ";"))
	values.Add("subject", subject)
	values.Add("body", body)
	resp, err := http.PostForm("http://127.0.0.1:8000/email", values)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("body: %w", err)
	}

	log.Println(string(buf))

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("invalid status: %s", resp.Status)
	}

	return nil
}
