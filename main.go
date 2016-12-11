package main

import (
	"bufio"
	"compress/gzip"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/blevesearch/bleve"
	"gopkg.in/yaml.v2"
)

func main() {
	switch os.Args[1] {
	case "process":
		process()
	case "search":
		results := search("learnprogramming", os.Args[2])
		fmt.Println(results)
	case "serve":
		serve(":8000")
	}
}

func getState() map[string]string {
	state := make(map[string]string)
	contents, err := ioutil.ReadFile("bleve/state.yaml")
	if err != nil {
		perr, ok := err.(*os.PathError)
		if !ok || perr.Err.Error() != "no such file or directory" {
			panic(err)
		}
		mkdir("bleve")
		contents, err = yaml.Marshal(state)
		if err != nil {
			panic(err)
		}
		ioutil.WriteFile("bleve/state.yaml", contents, 0644)
	} else {
		yaml.Unmarshal(contents, &state)
	}
	return state
}

func updateState(state map[string]string) {
	contents, err := yaml.Marshal(state)
	if err != nil {
		panic(err)
	}
	ioutil.WriteFile("bleve/state.yaml", contents, 0644)
}

func process() {
	state := getState()
	fmt.Println("state:", state["learnprogramming"])

	matches, err := filepath.Glob("/home/raylu/irclogs/learnprogramming/????-??-??.gz")
	if err != nil {
		panic(err)
	}
	index := getIndex("learnprogramming")
	for _, path := range matches {
		date := filepath.Base(path)
		date = date[:len(date)-3] // trim ".gz"
		if date <= state["learnprogramming"] {
			continue
		}
		fmt.Println(path)

		batch := index.NewBatch()
		err := processFile(batch, path, date)
		if err != nil {
			panic(err)
		}
		err = index.Batch(batch)
		if err != nil {
			panic(err)
		}
		state["learnprogramming"] = date
		updateState(state)
	}
	fmt.Println("done!")
}

func processFile(batch *bleve.Batch, filepath, date string) error {
	t, err := time.Parse("2006-01-02", date)
	if err != nil {
		return err
	}

	rawReader, err := os.Open(filepath)
	if err != nil {
		return err
	}
	defer rawReader.Close()
	gzipReader, err := gzip.NewReader(rawReader)
	if err != nil {
		return err
	}
	defer gzipReader.Close()

	scanner := bufio.NewScanner(gzipReader)
	lineNumber := 0
	for scanner.Scan() {
		line := scanner.Text()
		lineNumber++
		processLine(batch, date, t, lineNumber, line)
	}
	err = scanner.Err()
	return err
}

func processLine(batch *bleve.Batch, date string, dt time.Time, lineNumber int, line string) {
	if strings.HasPrefix(line, "--- ") { // --- Log opened/closed, Day changed
		return
	}
	// line should be one of
	// "12:34 <nick> msg"
	// "12:34  * nick action"
	// "12:34 -!- nick ..."
	if line[2] != ':' {
		panic("unexpected line: " + line)
	}
	if line[6:9] == "-!-" {
		return
	}
	var text string
	if line[6] == '<' {
		nickEnd := strings.IndexByte(line[7:], '>')
		if nickEnd == -1 {
			panic("unexpected line: " + line)
		}
		text = line[nickEnd+9:]
	} else if line[6:9] == " * " {
		nickEnd := strings.IndexByte(line[10:], ' ')
		if nickEnd == -1 {
			panic("unexpected line: " + line)
		}
		text = line[nickEnd+11:]
	} else {
		panic("unexpected line: " + line)
	}

	id := fmt.Sprintf("%s:%d", date, lineNumber)
	msg := ircMsg{Dt: dt, Text: text}
	batch.Index(id, msg)
}

func mkdir(dir string) {
	err := os.Mkdir(dir, 0755)
	if err != nil {
		perr, ok := err.(*os.PathError)
		if !ok || perr.Err.Error() != "file exists" {
			panic(err)
		}
	}
}
