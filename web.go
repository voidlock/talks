package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"text/template"
)

type File struct {
	Name string
	Path string
}

func (f File) Content() (string, error) {
	content, err := readContent(f.Path)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf(`{{define "%s"}}%s{{end}}`, f.Name, content), nil
}

func readContent(p string) (string, error) {
	info, err := os.Stat(p)
	if err != nil {
		if os.IsNotExist(err) {
			return "", err
		}
	}

	if info.IsDir() {
		return "", fmt.Errorf(p, "is a directory")
	}

	b, err := ioutil.ReadFile(p)
	return string(b), err
}

func serveSlides(w http.ResponseWriter, r *http.Request) {
	dir, file := path.Split(r.URL.Path)
	dir = path.Base(dir)
	file = path.Base(file)
	if dir == "/" {
		dir = "layout"
	}

	layout := File{dir, path.Join("slides", dir+".html")}
	slides := File{"markdown", path.Join("slides", file+".md")}

	t, err := parseFiles(layout, slides)
	if err != nil {
		http.NotFound(w, r)
	}

	if err := t.ExecuteTemplate(w, layout.Name, nil); err != nil {
		log.Println(err.Error())
		http.Error(w, http.StatusText(500), 500)
	}
}

func parseFiles(files ...File) (*template.Template, error) {
	var tmpl *template.Template

	for _, file := range files {
		if tmpl == nil {
			tmpl = template.New(file.Name)
		}
		content, err := file.Content()
		if err != nil {
			return nil, err
		}

		_, err = tmpl.Parse(content)
		if err != nil {
			return nil, err
		}

	}
	return tmpl, nil
}

func serveIndex(w http.ResponseWriter, r *http.Request) {
	http.NotFound(w, r)
}

func main() {
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	http.Handle("/slides/", http.StripPrefix("/slides", http.HandlerFunc(serveSlides)))
	http.HandleFunc("/", serveIndex)

	port := os.Getenv("PORT")
	log.Println("Listening on port", port)
	http.ListenAndServe(":"+port, nil)
}
