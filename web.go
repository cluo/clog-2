package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"path"
)

func serve(addr string) {
	http.Handle("/", http.HandlerFunc(webIndex))
	http.Handle("/log/", http.HandlerFunc(webLog))
	http.Handle("/search", http.HandlerFunc(webSearch))
	err := http.ListenAndServe(addr, nil)
	if err != nil {
		panic(err)
	}
}

func webIndex(w http.ResponseWriter, req *http.Request) {
	http.ServeFile(w, req, "web"+req.URL.Path)
}

func webLog(w http.ResponseWriter, req *http.Request) {
	src, err := ioutil.ReadFile("web/log.html")
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
	log, err := openLog(relpath)
	if err != nil {
		if lerr, ok := err.(*logNotFound); ok {
			http.Error(w, lerr.Error(), 404)
		} else {
			http.Error(w, err.Error(), 500)
		}
		return
	}
	defer log.Close()

	context := struct {
		Relpath string
		LogIter chan logline
	}{
		Relpath: relpath,
		LogIter: log.Iter(),
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
