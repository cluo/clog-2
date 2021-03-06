package main

import (
	"bufio"
	"compress/gzip"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/blevesearch/bleve"
	"gopkg.in/yaml.v2"
)

func main() {
	config = getConfig()

	switch os.Args[1] {
	case "process":
		process()
	case "search":
		results, err := search(os.Args[2], os.Args[3])
		if err != nil {
			fmt.Println(err)
		} else {
			for _, hit := range results {
				fmt.Println(hit)
			}
		}
	case "serve":
		go func() {
			for range time.NewTicker(1 * time.Hour).C {
				process()
			}
		}()
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
	log.Println("state:", state)

	for _, channel := range config.Channels {
		processChannel(state, channel)
	}
	log.Println("done!")
}

func processChannel(state map[string]string, channel string) {
	log.Println("processing", channel)
	matches, err := filepath.Glob(filepath.Join(config.Irclogs, channel, "????-??-??.gz"))
	if err != nil {
		panic(err)
	}
	index := getIndex(channel)
	for _, path := range matches {
		date := filepath.Base(path)
		date = date[:len(date)-3] // trim ".gz"
		if date <= state[channel] {
			continue
		}
		log.Println(path)

		batch := index.NewBatch()
		err := processFile(batch, path, date)
		if err != nil {
			panic(err)
		}
		err = index.Batch(batch)
		if err != nil {
			panic(err)
		}
		state[channel] = date
		updateState(state)
	}
}

func processFile(batch *bleve.Batch, filepath, date string) error {
	dt, err := time.Parse("2006-01-02", date)
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
	lastWindow := -1
	windowText := ""
	for scanner.Scan() {
		line := scanner.Text()
		sixHourWindow, text := parseLine(line)
		if sixHourWindow == -1 {
			continue
		} else if sixHourWindow == lastWindow || lastWindow == -1 {
			windowText += "\n" + text
			lastWindow = sixHourWindow
		} else {
			id := fmt.Sprintf("%s:%d", date, lastWindow)
			msg := ircMsg{Dt: dt, Text: windowText}
			batch.Index(id, msg)

			windowText = text
			lastWindow = sixHourWindow
		}
	}
	err = scanner.Err()
	if lastWindow != -1 {
		id := fmt.Sprintf("%s:%d", date, lastWindow)
		msg := ircMsg{Dt: dt, Text: windowText}
		batch.Index(id, msg)
	}
	return err
}

func parseLine(line string) (int, string) {
	if strings.HasPrefix(line, "--- ") { // --- Log opened/closed, Day changed
		return -1, ""
	}
	// line should be one of
	// "12:34 <nick> msg"
	// "12:34  * nick action"
	// "12:34 -!- nick ..."
	// "12:34 -nick:#channel- notice"
	// "12:34 nick [mask] requested CTCP ..."
	if line[2] != ':' {
		panic("unexpected line: " + line)
	}
	if line[6] == '-' {
		return -1, ""
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
	} else if strings.Index(line, "requested CTCP") == -1 && strings.Index(line, "requested unknown CTCP") == -1 {
		panic("unexpected line: " + line)
	}

	var hour int
	fmt.Sscanf(line[:2], "%d", &hour)
	return hour / 6, text
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
