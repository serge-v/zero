package zero

import (
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/serge-v/zero/server"
)

func TestDeploy(t *testing.T) {
	if false {
		ts := httptest.NewServer(http.HandlerFunc(server.HandleCompileAndDeployRequest))
		defer ts.Close()
		durl = ts.URL
	}

	log.SetFlags(log.LstdFlags | log.Lshortfile)

	if err := os.Chdir("cmd/testapp"); err != nil {
		t.Fatal(err)
	}

	err := Deploy()
	if err != nil {
		t.Fatal(err)
	}
}
