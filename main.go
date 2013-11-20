package main

import (
	m "github.com/paulbellamy/mango"
	"io/ioutil"
	"os"
)

var index []byte

func App(e m.Env) (m.Status, m.Headers, m.Body) {
	//return 200, nil, m.Body(index)
	return 200, nil, m.Body(index)
}

func main() {

	_, err := os.Stat("index.html")
	if err != nil {
		panic(err)
	}

	index, err = ioutil.ReadFile("index.html")

	if err != nil {
		panic(err)
	}

	app := m.Stack{}

	app.Address = ":8080"
	app.Run(App)
}
