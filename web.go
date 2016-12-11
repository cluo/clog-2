package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
)

var html []byte

func serve(addr string) {
	var err error
	html, err = ioutil.ReadFile("index.html")
	if err != nil {
		panic(err)
	}

	http.Handle("/", http.HandlerFunc(webIndex))
	http.Handle("/search", http.HandlerFunc(webSearch))
	err = http.ListenAndServe(addr, nil)
	if err != nil {
		panic(err)
	}
}

func webIndex(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(html)
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
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Write(output)
}
