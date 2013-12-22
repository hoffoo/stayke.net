package homepage

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"strings"
	"testing"
	"sync"
)

const content = `404.html
about.html
bg.png
cfooter.html
code.css
codepage.html
contact.html
index.html
main.css
projpage.html`

func TestWebApp(t *testing.T) {

	setup()

	pageUris := strings.Split(content, "\n")

	c := make(chan bool)

	for _, uri := range pageUris {

		u := uri
		go func() {
			var failure bool

			resp, err := http.Get("http://localhost:8999/" + u)
			if err != nil {
				t.Log(err)
			}

			if resp.StatusCode != 200 {
				t.Log(resp.Request.URL.Path + " failed")
				failure = true
			}

			c <- failure
		}()
	}

	respc := 0
	for {
		failure := <-c

		if failure {
			t.Fail()
		}

		respc++
		if respc == len(pageUris) {
			break
		}
	}
}

func TestProjectWebAppReq(t *testing.T) {

	setup()

	p := Projects{}
	filepath.Walk("code", p.Walker)

	reqC := 0
	c := make(chan bool)

	for proj, files := range p {

		reqC += len(files)

		proj := proj
		files := files

		go func() {
			for _, f := range files {
				resp, err := http.Get(fmt.Sprintf("http://localhost:8999/project/%s/%s", proj, f))

				if err != nil || resp.StatusCode != 200 {
					t.Log(fmt.Sprintf("code %d: /project/%s/%s", resp.StatusCode, proj, f))
					c <- true
				} else {
					c <- false
				}
			}
		}()
	}

	for {
		failed := <-c

		if (failed) {
			t.Fail()
		}

		reqC--
		if reqC == 0 {
			break
		}
	}
}

func TestLoadProjects(t *testing.T) {

	setup()

	p := Projects{}

	filepath.Walk("code", p.Walker)

	if len(p) == 0 {
		t.Fatal("didnt walk shit")
	}
}

func TestLoadPages(t *testing.T) {

	setup()

	ps := Pages{}

	filepath.Walk("html", ps.Walker)

	expectedPages := strings.Split(content, "\n")

	if len(ps) != len(expectedPages) {
		t.Log(content)
		t.Fatal(fmt.Sprintf("walked %d expected %d", len(ps), len(expectedPages)))
	}
}

var setupOnce sync.Once

func setup() {
	setupOnce.Do(startWebAppAndWait)
}

// starts the web app and waits a request to complete
func startWebAppAndWait() {

	go StartHTTP("localhost:8999")

	logBuff := bytes.Buffer{}
	log.SetOutput(&logBuff)

	// wait for app to start
	for {
		_, err := http.Get("http://localhost:8999/")
		if err == nil {
			break
		}
	}
}
