package main

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
)

var html []byte

func serve(addr string) {
	var err error
	if !config.RereadHtml {
		html, err = ioutil.ReadFile("index.html")
		if err != nil {
			panic(err)
		}
	}

	http.Handle("/", http.HandlerFunc(webIndex))
	http.Handle("/style.css", http.HandlerFunc(webCss))
	http.Handle("/log/", http.HandlerFunc(webLog))
	http.Handle("/search", http.HandlerFunc(webSearch))
	err = http.ListenAndServe(addr, nil)
	if err != nil {
		panic(err)
	}
}

func webIndex(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if config.RereadHtml {
		http.ServeFile(w, req, "index.html")
	} else {
		w.Write(html)
	}
}

func webCss(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/css; charset=utf-8")
	http.ServeFile(w, req, "style.css")
}

func webLog(w http.ResponseWriter, req *http.Request) {
	src, err := ioutil.ReadFile("log.html")
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	tpl, err := template.New("log").Parse(string(src))
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	relpath := req.URL.Path[5:] // strip "/log/"
	relpath = path.Clean(relpath)
	filepath := fmt.Sprintf("%s/%s.gz", config.Irclogs, relpath)

	var reader io.ReadCloser
	rawReader, err := os.Open(filepath)
	if err != nil {
		perr, ok := err.(*os.PathError)
		if !ok || perr.Err.Error() != "no such file or directory" {
			http.Error(w, err.Error(), 500)
			return
		}
		// try again without .gz
		reader, err = os.Open(fmt.Sprintf("%s/%s", config.Irclogs, relpath))
		if err != nil {
			http.Error(w, fmt.Sprintf("couldn't find \"%s\"", req.URL.Path[5:]), 404)
			return
		}
	} else {
		defer rawReader.Close()
		reader, err = gzip.NewReader(rawReader)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
	}
	defer reader.Close()
	logBytes, err := ioutil.ReadAll(reader)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	context := struct {
		Relpath string
		Log     string
	}{
		Relpath: relpath,
		Log:     string(logBytes),
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err = tpl.Execute(w, context)
	if err != nil {
		fmt.Println(err)
	}
}

func webSearch(w http.ResponseWriter, req *http.Request) {
	channel := req.URL.Query().Get("channel")
	q := req.URL.Query().Get("q")
	if channel == "" || q == "" {
		http.Error(w, "missing channel or q", 400)
		return
	}
	results := search(channel, q)
	output, err := json.MarshalIndent(results, "", "\t")
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Write(output)
}
