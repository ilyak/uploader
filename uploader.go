// (c) 2018 Ilya Kaliman
package main

import (
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

const formTemplate = `<!DOCTYPE html>
<html>
<head>
  <title>File Upload</title>
  <style>body { font-family: sans-serif; }</style>
</head>
<body>
  <form action="/upload" method="post" enctype="multipart/form-data">
    <input type="file" name="files" multiple />
    <input type="submit" value="Upload" />
  </form>
  <p>{{len .}} file(s)</p>
  <ul>
    {{range .}}<li><a href="/files/{{.}}">{{.}}</a></li>{{end}}
  </ul>
</body>
</html>`

const resultTemplate = `<!DOCTYPE html>
<html>
<head>
  <title>File Upload</title>
  <style>body { font-family: sans-serif; }</style>
</head>
<body>
  <p>Result: {{.}}; <a href="/">go back</a></p>
</body>
</html>`

func handler(w http.ResponseWriter, r *http.Request) {
	files, err := ioutil.ReadDir("files")
	if err != nil {
		log.Fatal(err)
	}
	var list []string
	for _, file := range files {
		list = append(list, file.Name())
	}
	t := template.Must(template.New("page").Parse(formTemplate))
	if err := t.Execute(w, list); err != nil {
		log.Fatal(err)
	}
}

func processReq(r *http.Request) string {
	if err := r.ParseMultipartForm(1 << 28); err != nil {
		return err.Error()
	}
	defer r.MultipartForm.RemoveAll()
	count := 0
	for _, list := range r.MultipartForm.File {
		for i := range list {
			if len(list[i].Filename) == 0 {
				continue
			}
			file, err := list[i].Open()
			if err != nil {
				return err.Error()
			}
			defer file.Close()
			path := "files/" + list[i].Filename
			if _, err := os.Stat(path); !os.IsNotExist(err) {
				return path + " exists"
			}
			out, err := os.Create(path)
			if err != nil {
				return err.Error()
			}
			defer out.Close()
			if _, err = io.Copy(out, file); err != nil {
				return err.Error()
			}
			count++
		}
	}
	if count == 0 {
		return "no files selected"
	}
	return fmt.Sprintf("successfully uploaded %d file(s)", count)
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	t := template.Must(template.New("page").Parse(resultTemplate))
	if err := t.Execute(w, processReq(r)); err != nil {
		log.Fatal(err)
	}
}

func main() {
	addr := "127.0.0.1:8080"
	if len(os.Args) > 1 {
		addr = os.Args[1]
	}
	os.Mkdir("files", os.ModePerm)
	http.HandleFunc("/", handler)
	http.HandleFunc("/upload", uploadHandler)
	fs := http.FileServer(http.Dir("files/"))
	http.Handle("/files", http.StripPrefix("/files", fs))
	fmt.Println("Listening on " + addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
