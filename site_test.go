package homepage

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

const content = `404.html
about.html
bg.png
robots.txt
code.css
codepage.html
contact.html
index.html
main.css
project.html`

func TestWebApp(t *testing.T) {

	setup(true)

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
				t.Fail()
				failure = true
			}

			c <- failure
		}()
	}

	respc := 0
	for {
		<-c

		respc++
		if respc == len(pageUris) {
			break
		}
	}
}

// test that we are sending back stuff that can be cached
func TestConditionalGet(t *testing.T) {

	setup(true)

	client := &http.Client{}
	req, err := http.NewRequest("GET", "http://localhost:8999/bg.png", nil)

	if err != nil {
		t.Fatal(err)
	}

	// req.Headers.Add()

	// make the request to bg.png
	resp, err := client.Do(req)

	if err != nil {
		t.Fatal(err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Fatal("didnt return 200 on img req")
	}

	if resp.Header.Get("Last-Modified") == "" || resp.Header.Get("Expires") == "" {
		t.Fatal("didnt send back cache control headers")
	}

	req, err = http.NewRequest("GET", "http://localhost:8999/bg.png", nil)

	mod := time.Now().Format(http.TimeFormat)
	req.Header.Add("If-Modified-Since", mod)

	resp, err = client.Do(req)

	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != 304 || resp.Header.Get("Last-Modified") == "" || resp.Header.Get("Expires") == "" {
		t.Fatal("didnt send back cache control headers or status")
	}

	defer resp.Body.Close()
}

func TestProjectWebAppReq(t *testing.T) {

	setup(true)

	p := Projects{}
	filepath.Walk("projects", p.Walker)

	reqC := 0
	c := make(chan bool)

	for proj, files := range p {

		proj := proj
		files := files

		reqC += len(files)
		//reqC += len(files) + 1

		//		resp, err := http.Get(fmt.Sprintf("http://localhost:8999/project/%s", proj))
		//
		//		// check all files and req to project /
		//		if err != nil || resp.StatusCode != 200 {
		//			t.Log("error opening / for project: " + proj)
		//			c <- true
		//		}

		go func() {

			for _, f := range files {
				resp, err := http.Get(fmt.Sprintf("http://localhost:8999/project/%s/%s", proj, f))

				if err != nil || resp.StatusCode != 200 {
					t.Log(fmt.Sprintf("projects %d: /project/%s/%s", resp.StatusCode, proj, f))
					t.Fail()
					c <- true
				} else {
					c <- false
				}
			}
		}()
	}

	for {
		<-c

		reqC--
		if reqC == 0 {
			break
		}
	}
}

func TestProjectCodeReq(t *testing.T) {

	setup(true)

	proj := "xxtail"
	f := "main.go.html"

	resp, err := http.Get(fmt.Sprintf("http://localhost:8999/code/%s/%s", proj, f))

	if resp.StatusCode != 200 || err != nil {
		t.Fatal("bad response from /code")
	}

	resp, err = http.Get(fmt.Sprintf("http://localhost:8999/code/BADPROJECT/%s", f))

	if resp.StatusCode != 404 || err != nil {
		t.Fatal("expected 404 on bad project in /code")
	}

	resp, err = http.Get(fmt.Sprintf("http://localhost:8999/code/xxtail"))

	if resp.StatusCode != 404 || err != nil {
		t.Fatal("expected 404 on bad project in /code")
	}

	resp, err = http.Get(fmt.Sprintf("http://localhost:8999/code/%s/%s", proj, f))

	if resp.StatusCode != 200 || err != nil {
		t.Fatal("expected to get file on /code")
	}
}

func TestMustachedNavAndCCS(t *testing.T) {

	setup(true)

	ps := Pages{}

	filepath.Walk("html", ps.Walker)

	p := Projects{}
	filepath.Walk("projects", p.Walker)

	mustacheTest := func(file, pattern, newdata string) {
		e := ps.Mustache(file, pattern, newdata)

		if e != nil || bytes.Index(ps[file], []byte(newdata)) == -1 {
			t.Logf("mustache failed for file [%s] pattern [%s]", file, pattern)
			t.Fail()
		}
	}

	mustacheTest("html/index.html", "{{{nav}}}", "")
	mustacheTest("html/project.html", "{{{nav}}}", "")
	mustacheTest("html/project.html", "{{{projectnav}}}", "")
	mustacheTest("html/contact.html", "{{{nav}}}", "")
	mustacheTest("html/about.html", "{{{nav}}}", "")
	mustacheTest("html/codepage.html", "{{{projectnav}}}", "")
	mustacheTest("html/codepage.html", "{{{filenav}}}", "")
	mustacheTest("html/codepage.html", "{{{url}}}", "")
}

func TestMustacheFunc(t *testing.T) {

	setup(true)

	ps := Pages{}

	filepath.Walk("html", ps.Walker)

	err := ps.Mustache("dfsfadss", "", "")

	if err == nil {
		t.Fatal("expected result: err on not finding bogus html file")
	}

	testPage := "html/project.html"

	beforeIdx := bytes.Index(ps[testPage], []byte("{{{projectnav}}}"))
	if beforeIdx == -1 {
		t.Fatal("no replace str")
	}

	err = ps.Mustache(testPage, "{{{projectnav}}}", "fail")
	if err != nil {
		t.Fatal("failed replacing in Mustache")
	}

	beforeIdx = bytes.Index(ps[testPage], []byte("{{{projectnav}}}"))
	if beforeIdx != -1 {
		t.Fatal("failed replacing in Mustache")
	}
}

func TestLoadProjects(t *testing.T) {

	setup(true)

	p := Projects{}

	filepath.Walk("projects", p.Walker)

	if len(p) == 0 {
		t.Fatal("didnt walk shit")
	}
}

func TestLoadPages(t *testing.T) {

	setup(true)

	ps := Pages{}

	filepath.Walk("html", ps.Walker)

	expectedPages := strings.Split(content, "\n")

	if len(ps) != len(expectedPages) {
		t.Log(content)
		t.Fatal(fmt.Sprintf("walked %d expected %d", len(ps), len(expectedPages)))
	}
}

func TestMimeType(t *testing.T) {

	setup(true)

	url := "http://localhost:8999/main.css"

	req, err := http.Get(url)

	if err != nil {
		t.Fatal(err)
	}

	if req.Header.Get("Content-Type") != "text/css; charset=utf-8" {
		t.Fatal("didnt get proper mime type " + req.Header.Get("Content-Type"))
	}
	req.Body.Close()

	url = "http://localhost:8999/bg.png"

	req, err = http.Get(url)

	if err != nil {
		t.Fatal(err)
	}

	if req.Header.Get("Content-Type") != "image/png" {
		t.Fatal("didnt get proper mime type " + req.Header.Get("Content-Type"))
	}
	req.Body.Close()
}

func TestResumeForward(t *testing.T) {

	setup(true)

	url := "http://localhost:8999/resume/"
	noredir := &http.Transport{}
	req, err := http.NewRequest("GET", url, nil)

	if err != nil {
		t.Fatal(err)
	}

	resp, err := noredir.RoundTrip(req)

	if err != nil {
		t.Fatal(err)
	}

	if resp.Header.Get("Location") == "" || resp.StatusCode != 301 {
		t.Fatal("didnt forward /resume request")
	}

	resp.Body.Close()
}

var setupOnce sync.Once

func setup(silenceLog bool) {

	if silenceLog {
		nullBuff := NullBuff{}
		log.SetOutput(&nullBuff)
	} else {
		log.SetOutput(os.Stdout)
	}

	setupOnce.Do(startWebAppAndWait)
}

// starts the web app and waits a request to complete
func startWebAppAndWait() {

	go StartHTTP("localhost:8999")

	// wait for app to start
	for {
		_, err := http.Get("http://localhost:8999/")
		if err == nil {
			break
		}
	}
}

type NullBuff bytes.Buffer

func (nb *NullBuff) Write(p []byte) (n int, err error) {
	// do nothing
	return n, nil
}
