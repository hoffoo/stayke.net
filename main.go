package main

import (
	"github.com/paulbellamy/mango"
	"io/ioutil"
	"regexp"
)

var debug bool = true

// could just do a static handler but where is the fun in that
// also my vps has a slow spinning disk and enough ram
var index []byte
var about []byte
var contact []byte
var css []byte
var bg []byte

func cacheFiles() {
	var ierr, aerr, cerr, csserr, bgerr error

	index, ierr = ioutil.ReadFile("html/index.html")
	about, aerr = ioutil.ReadFile("html/about.html")
	contact, cerr = ioutil.ReadFile("html/contact.html")
	css, csserr = ioutil.ReadFile("html/main.css")
	bg, bgerr = ioutil.ReadFile("html/bg.png")

	for _, err := range []error{ierr, aerr, cerr, csserr, bgerr} {
		if err != nil {
			panic(err)
		}
	}
}

func Index(e mango.Env) (mango.Status, mango.Headers, mango.Body) {
	cacheFiles()
	return 200, nil, mango.Body(index)
}

func About(e mango.Env) (mango.Status, mango.Headers, mango.Body) {
	cacheFiles()
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

func FourOhFour(e mango.Env) (mango.Status, mango.Headers, mango.Body) {
	noPage, err := ioutil.ReadFile("html/404.html")

	if err != nil {
		noPage = []byte("page cannot be found")
	}

	return 404, nil, mango.Body(noPage)
}

// regex to check the url against and grab project name
var pregex *regexp.Regexp

func Projects(e mango.Env) (mango.Status, mango.Headers, mango.Body) {
	path := e.Request().URL.Path
	p := pregex.FindStringSubmatch(path)

	if p == nil {
		// bad url, leave this
		return FourOhFour(e)
	}

	return 404, nil, mango.Body([]byte(p[1] + p[2]))
}

func main() {

	cacheFiles()

	// first match is the project name, second relative path of the file
	pregex = regexp.MustCompile("/code/([\\w-_\\d]+)(/[\\w-_\\d\\/\\.]+)?$")

	app := mango.Stack{}

	app.Address = ":8080"

	r := map[string]mango.App{
		"/$|index": app.Compile(Index),
		"about":    app.Compile(About),
		"contact":  app.Compile(Contact),
		"main.css": app.Compile(Css),
		"bg.png":   app.Compile(Bg),
		"/code/*":  app.Compile(Projects),
	}

	app.Middleware(mango.Routing(r))
	app.Run(FourOhFour)
}
