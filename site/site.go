package site

import (
    "bytes"
    "fmt"
    M "github.com/hoffoo/marin"
    "io/ioutil"
    "os"
    "path"
    "text/template"
)

type Page struct {
    Content string
}

func GetPages() (pages map[string]*bytes.Buffer, documents map[string]*bytes.Buffer) {

    pageTpl := template.New("page")
    pageTpl = packTemplate(pageTpl, "page")
    pageTpl = packTemplate(pageTpl, "head")
    pageTpl = packTemplate(pageTpl, "nav")

    homePage := getPage("home")
    aboutPage := getPage("about")
    contactPage := getPage("contact")

    homeBuf := bytes.Buffer{}
    aboutBuf := bytes.Buffer{}
    contactBuf := bytes.Buffer{}

    pageTpl.Execute(&homeBuf, homePage)
    pageTpl.Execute(&aboutBuf, aboutPage)
    pageTpl.Execute(&contactBuf, contactPage)

    pages = map[string]*bytes.Buffer{
        "/home":    &homeBuf,
        "/about":   &aboutBuf,
        "/contact": &contactBuf,
    }

    documents = map[string]*bytes.Buffer{
        "/main.css":   readAll("main.css"),
        "/bg.png":     readAll("bg.png"),
        "/robots.txt": readAll("robots.txt"),
    }

    return
}

func getPage(name string) (page *Page) {
    cwd, err := os.Getwd()
    M.PANIC_ON_ERR(err)

    f, err := os.Open(fmt.Sprintf("%s/%s.html", path.Join(cwd, "templates"), name))
    M.PANIC_ON_ERR(err)
    defer f.Close()

    content, err := ioutil.ReadAll(f)
    M.PANIC_ON_ERR(err)

    return &Page{
        Content: string(content),
    }
}

func readAll(file string) *bytes.Buffer {

    cwd, err := os.Getwd()
    M.PANIC_ON_ERR(err)

    f, err := os.Open(fmt.Sprintf("%s/%s", path.Join(cwd, "templates"), file))
    M.PANIC_ON_ERR(err)
    defer f.Close()

    content, err := ioutil.ReadAll(f)
    return bytes.NewBuffer(content)
}

func packTemplate(tplPack *template.Template, name string) (tpl *template.Template) {

    cwd, err := os.Getwd()
    M.PANIC_ON_ERR(err)

    f, err := os.Open(fmt.Sprintf("%s/%s.html", path.Join(cwd, "templates"), name))
    M.PANIC_ON_ERR(err)
    defer f.Close()

    html, err := ioutil.ReadAll(f)
    M.PANIC_ON_ERR(err)

    tpl, err = tplPack.Parse(string(html))
    M.PANIC_ON_ERR(err)

    return
}
