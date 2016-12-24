package main

import (
	"encoding/json"
	"html/template"
	"io/ioutil"
	"log"
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
	if req.URL.Path == "/" {
		tpl, err := loadTemplate("index")
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		context := struct {
			Channels []string
		}{
			Channels: config.Channels,
		}
		err = tpl.Execute(w, context)
		if err != nil {
			log.Println(err)
		}
	} else {
		http.ServeFile(w, req, "web"+req.URL.Path)
	}
}

func webLog(w http.ResponseWriter, req *http.Request) {
	tpl, err := loadTemplate("log")
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	relpath := req.URL.Path[5:] // strip "/log/"
	relpath = path.Clean(relpath)
	lf, err := openLog(relpath)
	if err != nil {
		if lerr, ok := err.(*logNotFound); ok {
			http.Error(w, lerr.Error(), 404)
		} else {
			http.Error(w, err.Error(), 500)
		}
		return
	}
	defer lf.Close()

	context := struct {
		Relpath string
		LogIter chan logline
	}{
		Relpath: relpath,
		LogIter: lf.Iter(),
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err = tpl.Execute(w, context)
	if err != nil {
		log.Println(err)
	}
}

func webSearch(w http.ResponseWriter, req *http.Request) {
	channel := req.URL.Query().Get("channel")
	q := req.URL.Query().Get("q")
	if channel == "" || q == "" {
		http.Error(w, "missing channel or q", 400)
		return
	}

	results, err := search(channel, q)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	output, err := json.MarshalIndent(results, "", "\t")
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Write(output)
}

func loadTemplate(name string) (*template.Template, error) {
	src, err := ioutil.ReadFile("web/" + name + ".html")
	if err != nil {
		return nil, err
	}
	return template.New(name).Parse(string(src))
}
