package main

import (
    "bytes"
    "fmt"
    "github.com/hoffoo/stayke.net/site"
    "github.com/howeyc/fsnotify"
    "net/http"
)

var pageResp map[string]*bytes.Buffer
var assetResp map[string]*bytes.Buffer

func init() {
    pageResp = map[string]*bytes.Buffer{}
    assetResp = map[string]*bytes.Buffer{}
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
    pages, assets := site.GetDocuments()

    for url, doc := range assets {
        assetResp[url] = doc
    }

    for url, page := range pages {
        pageResp[url+".html"] = page
        pageResp[url] = page
    }
    pageResp["/"] = pageResp["/home"]
    pageResp["/index"] = pageResp["/home"]
    pageResp["/index.html"] = pageResp["/home"]

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

    if page, ok := pageResp[r.URL.Path]; ok {
        w.Header().Set("Content-Type", "text/html")
        w.Write(page.Bytes())
        return
    }

    if asset, ok := assetResp[r.URL.Path]; ok {
        w.Header().Set("Content-Type", "text/html")
        w.Write(asset.Bytes())
        return
    }

    w.WriteHeader(http.StatusNotFound)
    w.Write(assetResp["404"].Bytes())
}
