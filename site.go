package homepage

import (
	"github.com/paulbellamy/mango"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"fmt"
)

var (
	html Pages
	projects Projects
)

type (
	Projects map[string][]string
	Pages map[string][]byte
)

func StartHTTP(addr string) {

	html = Pages{}
	filepath.Walk("html", html.Walker)

	projects = Projects{}
	filepath.Walk("code", projects.Walker)

	app := mango.Stack{}
	app.Address = addr

	r := map[string]mango.App{
		"/project/*": app.Compile(Project),
		"/*":         app.Compile(Index),
	}

	app.Middleware(mango.Routing(r))
	app.Run(FourOhFour)
}

func Index(e mango.Env) (mango.Status, mango.Headers, mango.Body) {

	f := "html" + e.Request().URL.Path
	page, have := html[f]
	if !have {
		log.Printf("404 on: %s", f)
		return FourOhFour(e)
	}

	log.Printf("request for %s", f)
	return 200, nil, mango.Body(page)
}

func Project(e mango.Env) (mango.Status, mango.Headers, mango.Body) {

	p := e.Request().URL.Path
	fmt.Sprintf(p)

	return 200, nil, mango.Body([]byte("wtf"))
}

func FourOhFour(e mango.Env) (mango.Status, mango.Headers, mango.Body) {
	return 404, nil, mango.Body(html["404.html"])
}

func (p Projects) Walker(path string, info os.FileInfo, err error) error {

	// TODO no subdirs
	if info.IsDir() {
		level := strings.Count(path, "/")
		if level <= 1 {
			log.Printf("adding project: %s", path)
			return nil
		} else {
			return filepath.SkipDir
		}
	}

	// skip the first dir (code)
	// set second dir as the key (project name)
	// set the rest as the filepath, no subdirs
	spl := strings.Split(path, "/")
	p[spl[1]] = append(p[spl[1]], spl[2])
	log.Printf("added [%s] %s", spl[1], path)

	return nil
}

func (ps Pages) Walker(path string, info os.FileInfo, err error) error {

	// lame way to check if this is the root that was passed
	if info.IsDir() && strings.Index(path, "/") > -1 {
		return filepath.SkipDir
	}

	fileb, err := ioutil.ReadFile(path)
	if err != nil {
		log.Print(err)
	} else {
		ps[path] = fileb
		log.Print("cached: " + path)
	}

	return nil
}
