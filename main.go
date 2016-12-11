package main

import (
	"bufio"
	"compress/gzip"
	"fmt"
	"os"
	"strings"
)

type ircMsg struct {
	Text string
}

func main() {
	switch os.Args[1] {
	case "process":
		process()
	case "search":
		results := search("learnprogramming", os.Args[2])
		fmt.Println(results)
	}
}

func process() {
	err := processFile("/home/raylu/irclogs/learnprogramming/2013-07-18.gz", "learnprogramming", "2013-07-18")
	if err != nil {
		panic(err)
	}
	fmt.Println("done!")
}

func processFile(filepath, channel, date string) error {
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
		processLine(channel, date, lineNumber, line)
	}
	err = scanner.Err()
	return err
}

func processLine(channel, date string, lineNumber int, line string) {
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

	msg := ircMsg{Text: text}
	index := getIndex(channel)
	index.Index(fmt.Sprintf("%s:%d", date, lineNumber), msg)
}
