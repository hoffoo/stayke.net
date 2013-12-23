package homepage

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/paulbellamy/mango"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

var (
	html     Pages
	projects Projects
)

type (
	Projects map[string][]string
	Pages    map[string][]byte
)

func StartHTTP(addr string) {

	html = Pages{}
	filepath.Walk("html", html.Walker)

	projects = Projects{}
	filepath.Walk("projects", projects.Walker)

	projects.MakeNav()
	html.MakeNav()

	projects.CleanupFileLinks()

	app := mango.Stack{}
	app.Address = addr

	r := map[string]mango.App{
		"^/project/.*": app.Compile(Project),
		"^/code/.*":    app.Compile(ProjectCode),
		"^/.*":         app.Compile(Index),
	}

	app.Middleware(mango.Routing(r))
	app.Run(FourOhFour)
}

func Index(e mango.Env) (mango.Status, mango.Headers, mango.Body) {

	// urls with .html and without
	// handle / and /index
	f := e.Request().URL.Path

	if f == "/" || f == "/home" {
		f = "/index"
	}

	if strings.Index(f, ".") > -1 {
		f = "html" + f
	} else {
		f = "html" + f + ".html"
	}

	page, have := html[f]

	if !have {
		log.Printf("404 on: %s", f)
		return FourOhFour(e)
	}

	log.Printf("req for %s", f)
	return 200, nil, mango.Body(page)
}

func Project(e mango.Env) (mango.Status, mango.Headers, mango.Body) {

	path := strings.Split(e.Request().URL.Path, "/")

	if len(path) < 3 {
		log.Printf("invalid project request: %s", path)
		return FourOhFour(e)
	}

	var p, f string

	// accepts requests in the form project/file, project/, project
	// if no file send README.md.html
	p = path[2]
	if len(path) == 3 {
		f = "README.md.html"
	} else if f = path[3]; f == "" {
		f = "README.md.html"
	}

	projFiles, have := projects[p]

	if !have {
		log.Printf("req for invalid project: %s", p)
		return FourOhFour(e)
	}

	iframeUrl := fmt.Sprintf("/code/%s/%s", p, f)

	projectNav := make([]string, len(projFiles))
	for i, pfile := range projFiles {
		projectNav[i] = fmt.Sprintf(`<a href="/project/%s/%s.html">%s</a>`, p, pfile, pfile)
	}
	nav := strings.Join(projectNav, "\n")

	codePage := html["html/codepage.html"]
	codePage = bytes.Replace(codePage, []byte("{{{url}}}"), []byte(iframeUrl), 1)
	codePage = bytes.Replace(codePage, []byte("{{{filenav}}}"), []byte(nav), 1)

	log.Printf("req for project file [%s]: %s", p, f)

	return 200, nil, mango.Body(codePage)
}

func ProjectCode(e mango.Env) (mango.Status, mango.Headers, mango.Body) {

	pathSpl := strings.Split(e.Request().URL.Path, "/")

	if len(pathSpl) != 4 {
		log.Printf("no file for /code req: %s", e.Request().URL.Path)
		return FourOhFour(e)
	}

	p := pathSpl[2]
	f := pathSpl[3]

	_, have := projects[p]

	if !have {
		log.Printf("req for bad project: %s", p)
		return FourOhFour(e)
	}

	file, err := ioutil.ReadFile(fmt.Sprintf("projects/%s/%s", p, f))

	if err != nil {
		log.Printf("couldnt load /code file [%s]: %s", p, f)
		log.Print(err)
		return FourOhFour(e)
	}

	log.Printf("req for project code [%s]: %s", p, f)

	return 200, nil, mango.Body(file)
}

func FourOhFour(e mango.Env) (mango.Status, mango.Headers, mango.Body) {
	return 404, nil, mango.Body(html["html/404.html"])
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

	// skip the first dir (projects)
	// set second dir as the key (project name)
	// set the rest as the filepath
	spl := strings.SplitN(path, "/", 3)
	p[spl[1]] = append(p[spl[1]], spl[2])
	log.Printf("added [%s] %s", spl[1], path)

	return nil
}

func (p Projects) MakeNav() {

	var result []string
	for project, _ := range p {
		result = append(result, fmt.Sprintf(`<a href="%s">%s</a>`, "/project/"+project, project))
	}

	html.Mustache("html/project.html", "{{{projectnav}}}", strings.Join(result, "\n"))
	html.Mustache("html/codepage.html", "{{{projectnav}}}", strings.Join(result, "\n"))

	log.Printf("finished project nav")
}

// fix the file links and turn into links
func (p Projects) CleanupFileLinks() {
	for pr, fAr := range p {
		for i, f := range fAr {
			p[pr][i] = strings.Replace(f, ".html", "", 1)
		}
	}
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

// replace out {{{}} for special pages
func (ps Pages) Mustache(file, pattern, newdata string) error {

	page, have := ps[file]

	if !have {
		log.Printf("unable to mustache %s: no such file", file)
		return errors.New("unable to mustache: no such file")
	}

	newpage := bytes.Replace(page, []byte(pattern), []byte(newdata), 1)
	ps[file] = newpage

	if bytes.Equal(page, newpage) {
		log.Printf("didnt replace %s in %s", pattern, file)
		return errors.New("mustache didnt replace " + pattern + " in " + file)
	}

	return nil
}

func (ps Pages) MakeNav() {

	navPages := []string{
		"home",
		"about",
		"project",
		"resume",
		"contact",
	}

	for i, l := range navPages {
		navPages[i] = fmt.Sprintf(`<a href="/%s">%s</a>`, l, l)
	}

	nav := "<div id=\"nav\">\n" + strings.Join(navPages, "\n") + "</div>"

	// TODO check err here in case replace fails
	ps.Mustache("html/index.html", "{{{nav}}}", nav)
	ps.Mustache("html/about.html", "{{{nav}}}", nav)
	ps.Mustache("html/project.html", "{{{nav}}}", nav)
	ps.Mustache("html/contact.html", "{{{nav}}}", nav)
	ps.Mustache("html/contact.html", "{{{nav}}}", nav)
}
