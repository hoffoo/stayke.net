package main

import (
    "./site"
    "bytes"
    "fmt"
    "github.com/howeyc/fsnotify"
    "mime"
    "net/http"
)

var responses map[string]*bytes.Buffer

func init() {
    responses = map[string]*bytes.Buffer{}
}

func main() {

    // loop
    // 		homepage.GetRoutes()
    // 		go listen for file changes
    // 		ListenAndServe
    // 		on file change call init to reprocess everything
    // endloop

    watcher, err := fsnotify.NewWatcher()
    if err != nil {
        panic(err)
    }

    watcher.Watch("templates")

    http.HandleFunc("/", BufferHandlerFunc)

reload:
    pages, documents := site.GetPages()

    for url, doc := range documents {
        responses[url] = doc
    }

    for url, page := range pages {
        responses[url+".html"] = page
        responses[url] = page
    }
    responses["/"] = responses["/home"]
    responses["/index"] = responses["/home"]
    responses["/index.html"] = responses["/home"]

    fmt.Println("Reloading...")
    go http.ListenAndServe(":9999", nil)

watch:
    ev := <-watcher.Event
    if !ev.IsModify() {
        goto watch
    } else {
        goto reload
    }

}

func BufferHandlerFunc(w http.ResponseWriter, r *http.Request) {

    if doc, ok := responses[r.URL.Path]; ok {
        // i dont care if there is no exenstion, this is only to handle css files
        w.Header().Set("Content-Type", mime.TypeByExtension(r.URL.Path))
        w.Write(doc.Bytes())
    } else {
        w.Write([]byte("not found"))
    }
}
