package main

import (
	"bufio"
	"compress/gzip"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/blevesearch/bleve"
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

func process() {
	matches, err := filepath.Glob("/home/raylu/irclogs/learnprogramming/????-??-??.gz")
	if err != nil {
		panic(err)
	}
	index := getIndex("learnprogramming")
	for _, path := range matches {
		fmt.Println(path)
		date := filepath.Base(path)
		date = date[:len(date)-3] // trim ".gz"
		batch := index.NewBatch()
		err := processFile(batch, path, date)
		if err != nil {
			panic(err)
		}
		err = index.Batch(batch)
		if err != nil {
			panic(err)
		}
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
