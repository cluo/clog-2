package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

var html []byte

func serve(addr string) {
	fileInfo, err := os.Stat("index.html")
	if err != nil {
		panic(err)
	}
	reader, err := os.Open("index.html")
	if err != nil {
		panic(err)
	}
	defer reader.Close()
	html = make([]byte, fileInfo.Size())
	n, err := reader.Read(html)
	if err != nil {
		panic(err)
	} else if n != int(fileInfo.Size()) {
		panic(fmt.Sprintf("expected %d bytes, got %d", fileInfo.Size(), n))
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
