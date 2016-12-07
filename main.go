package main

import (
	"bufio"
	"compress/gzip"
	"fmt"
	"os"
	"strings"
)

func main() {
	fmt.Println("hi!")
	err := processFile("/home/raylu/irclogs/learnprogramming/2013-07-18.gz")
	if err != nil {
		panic(err)
	}
}

func processFile(filepath string) error {
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
	for scanner.Scan() {
		line := scanner.Text()
		processLine(line)
	}
	err = scanner.Err()
	return err
}

func processLine(line string) {
	if strings.HasPrefix(line, "--- Log ") { // --- Log opened/closed
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
	var msg string
	if line[6] == '<' {
		nickEnd := strings.IndexByte(line[7:], '>')
		if nickEnd == -1 {
			panic("unexpected line: " + line)
		}
		msg = line[nickEnd+9:]
	} else if line[6:9] == " * " {
		nickEnd := strings.IndexByte(line[10:], ' ')
		if nickEnd == -1 {
			panic("unexpected line: " + line)
		}
		msg = line[nickEnd+11:]
	} else {
		panic("unexpected line: " + line)
	}

	msg = strings.ToLower(msg)
	for _, word := range strings.Split(msg, " ") {
		if word == "" {
			continue
		}
		processWord(word)
	}
}

func processWord(word string) {
}
