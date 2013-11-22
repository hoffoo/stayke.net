package main

import (
	"bytes"
	"fmt"
	"github.com/paulbellamy/mango"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var debug bool = true

const (
	projDir = "html/code"
	codeDir = "src/"
)

func Projects(e mango.Env) (mango.Status, mango.Headers, mango.Body) {

	p, f, failed := resolvePath(e)
	if failed == true {
		return FourOhFour(e)
	}

	// TODO stat project dir

	// poor man's mustache
	body := bytes.Replace(cpage, []byte("{{{url}}}"), []byte(p+f+".html"), 1)
	body = bytes.Replace(body, []byte("{{{pnav}}}"), []byte(pnav), 1)

	pfiles, _ := fileNav[p]
	fnav := strings.Join(pfiles, "\n")

	body = bytes.Replace(body, []byte("{{{fnav}}}"), []byte(fnav), 1)
	return 200, nil, mango.Body(body)
}

func Code(e mango.Env) (mango.Status, mango.Headers, mango.Body) {
	p, f, failed := resolvePath(e)
	if failed == true {
		pregex := regexp.MustCompile("code/([\\w-_\\d]+).html")
		parr := pregex.FindStringSubmatch(e.Request().URL.Path)

		if parr == nil {
			return FourOhFour(e)
		} else {
			p = parr[1]
			f = "/README.md.html"
		}
	}

	fpath := "html/code/" + p + f

	fileb, err := ioutil.ReadFile(fpath)
	if err != nil {
		return FourOhFour(e)
	} else {
		return 200, nil, mango.Body(fileb)
	}

	return FourOhFour(e)
}

// resolves url path to project and file path
func resolvePath(e mango.Env) (project, file string, err bool) {
	p := e.Request().URL.Path

	// TODO optimize this
	// first match is the project name, second relative path of the file
	pregex := regexp.MustCompile("/(code|project)/([\\w-_\\d]+)?(/[\\w-_\\d\\/\\.]+)?$")
	pr := pregex.FindStringSubmatch(p)

	if pr == nil {
		// bad url, leave this
		return "", "", true
	}

	project = pr[2]
	file = pr[3]
	err = false

	return
}

// list of projects
func ProjectList(e mango.Env) (mango.Status, mango.Headers, mango.Body) {
	body := bytes.Replace(ppage, []byte("{{{pnav}}}"), []byte(pnav), 1)

	return 200, nil, mango.Body(body)
}

// could just do a static handler but where is the fun in that
// also my vps has a slow spinning disk and enough ram
// so cache the pages and code page partials
var (
	// main pages
	index   []byte
	about   []byte
	contact []byte
	css     []byte
	bg      []byte
	// projects page (index)
	ppage []byte
	// code page header and footer
	cpage  []byte
	cstyle []byte
)

func cacheFiles() {
	ioread(&index, "html/index.html")
	ioread(&about, "html/about.html")
	ioread(&contact, "html/contact.html")
	ioread(&css, "html/main.css")
	ioread(&bg, "html/bg.png")
	ioread(&cpage, "html/codepage.html")
	ioread(&cstyle, "html/code.css")
	ioread(&ppage, "html/projpage.html")
}

func ioread(buff *[]byte, path string) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}

	*buff = b
}

var (
	pnav    string
	fileNav map[string][]string
)

const (
	projLink = "<a href=\"/project/%s\">%s</a>\n"
	fileLink = "<a href=\"/project/%s/%s\">%s</a>\n"
)

// build the html for both the project header
// and file browser
func cacheProjects() {
	d, err := os.OpenFile(projDir, os.O_RDONLY, 0660)
	if err != nil {
		panic(err)
	}

	dfs, rerr := d.Readdir(-1)
	if rerr != nil {
		panic(rerr)
	}

	for _, f := range dfs {
		if f.IsDir() {
			pnav = pnav + fmt.Sprintf(projLink, f.Name(), f.Name())
		}
	}

	fileNav = make(map[string][]string)

	var root string
	walker := func(p string, info os.FileInfo, err error) error {
		if info.IsDir() {
			root = info.Name()
		} else {
			// TODO this is kind of retarded
			title := strings.Replace(info.Name(), ".html", "", 1)
			fileNav[root] = append(fileNav[root], fmt.Sprintf(fileLink, root, title, title))
		}

		return nil
	}

	for _, d := range dfs {
		if d.IsDir() {
			filepath.Walk(projDir+"/"+d.Name(), walker)
		}
	}
}

// TODO get rid of all these redundant functions
func Index(e mango.Env) (mango.Status, mango.Headers, mango.Body) {
	return 200, nil, mango.Body(index)
}

func About(e mango.Env) (mango.Status, mango.Headers, mango.Body) {
	return 200, nil, mango.Body(about)
}

func Contact(e mango.Env) (mango.Status, mango.Headers, mango.Body) {
	return 200, nil, mango.Body(contact)
}

func Css(e mango.Env) (mango.Status, mango.Headers, mango.Body) {
	return 200, nil, mango.Body(css)
}

func Bg(e mango.Env) (mango.Status, mango.Headers, mango.Body) {
	return 200, nil, mango.Body(bg)
}

func CodeCss(e mango.Env) (mango.Status, mango.Headers, mango.Body) {
	return 200, nil, mango.Body(cstyle)
}

func FourOhFour(e mango.Env) (mango.Status, mango.Headers, mango.Body) {
	noPage, err := ioutil.ReadFile("html/404.html")

	if err != nil {
		noPage = []byte("page cannot be found")
	}

	return 404, nil, mango.Body(noPage)
}

func main() {

	cacheFiles()
	cacheProjects()

	app := mango.Stack{}

	app.Address = ":8080"

	r := map[string]mango.App{
		"/$|index":      app.Compile(Index),
		"about":         app.Compile(About),
		"contact":       app.Compile(Contact),
		"main.css":      app.Compile(Css),
		"bg.png":        app.Compile(Bg),
		"code.css":      app.Compile(CodeCss),
		"/project(/)?$": app.Compile(ProjectList),
		"/project/*":    app.Compile(Projects),
		"/code/*":       app.Compile(Code),
	}

	app.Middleware(mango.Routing(r))
	app.Run(FourOhFour)
}
