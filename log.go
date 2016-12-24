package main

import (
	"bufio"
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

type logfile struct {
	reader    io.ReadCloser
	rawReader io.ReadCloser
	scanner   *bufio.Scanner
}

type logline struct {
	Number    int
	Timestamp string
	Text      string
}

var windows map[int]map[string]bool

func init() {
	windows = make(map[int]map[string]bool)
	for i := 0; i < 4; i++ {
		windows[i] = make(map[string]bool, 24/6)
		for j := i * 6; j < (i+1)*6; j++ {
			windows[i][fmt.Sprintf("%02d", j)] = true
		}
	}
}

func openLog(relpath string) (*logfile, error) {
	filepath := fmt.Sprintf("%s/%s.gz", config.Irclogs, relpath)
	var reader io.ReadCloser
	rawReader, err := os.Open(filepath)
	if err != nil {
		perr, ok := err.(*os.PathError)
		if !ok || perr.Err.Error() != "no such file or directory" {
			return nil, err
		}
		// try again without .gz
		reader, err = os.Open(fmt.Sprintf("%s/%s", config.Irclogs, relpath))
		if err != nil {
			return nil, &logNotFound{relpath: relpath}
		}
		rawReader = nil
	} else {
		reader, err = gzip.NewReader(rawReader)
		if err != nil {
			return nil, err
		}
	}
	scanner := bufio.NewScanner(reader)
	lf := &logfile{reader: reader, rawReader: rawReader, scanner: scanner}
	return lf, nil
}

func (lf *logfile) Close() {
	lf.reader.Close()
	if lf.rawReader != nil {
		lf.rawReader.Close()
	}
}

func (lf *logfile) Iter() chan logline {
	ch := make(chan logline)
	go func() {
		lineNo := 0
		for lf.scanner.Scan() {
			lineNo++
			line := lf.scanner.Text()
			if len(line) > 7 && line[2] == ':' {
				ch <- logline{lineNo, line[:5], line[6:]}
			} else {
				ch <- logline{lineNo, "", line}
			}
		}
		err := lf.scanner.Err()
		if err != nil {
			log.Println("logfile Iter:", err)
		}
		close(ch)
	}()
	return ch
}

func (lf *logfile) Line(sixHourWindow int, fragment string) (int, string) {
	hours := windows[sixHourWindow]
	lineNo := 0
	for lf.scanner.Scan() {
		lineNo++
		line := lf.scanner.Text()
		if len(line) <= 7 || line[2] != ':' {
			continue
		}
		if hours[line[:2]] && strings.Index(line, fragment) > -1 {
			return lineNo, line
		}
	}
	err := lf.scanner.Err()
	if err != nil {
		log.Println("logfile Line:", err)
	}
	return -1, ""
}

type logNotFound struct {
	relpath string
}

func (lnf *logNotFound) Error() string {
	return fmt.Sprintf("couldn't find \"%s\"", lnf.relpath)
}
