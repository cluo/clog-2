package main

import (
	"bufio"
	"compress/gzip"
	"fmt"
	"io"
	"os"
)

type logfile struct {
	reader    io.ReadCloser
	rawReader io.ReadCloser
	scanner   *bufio.Scanner
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

func (lf *logfile) Iter() chan string {
	ch := make(chan string)
	go func() {
		for lf.scanner.Scan() {
			ch <- lf.scanner.Text()
		}
		err := lf.scanner.Err()
		if err != nil {
			fmt.Println("logfile Iter:", err)
		}
		close(ch)
	}()
	return ch
}

type logNotFound struct {
	relpath string
}

func (lnf *logNotFound) Error() string {
	return fmt.Sprintf("couldn't find \"%s\"", lnf.relpath)
}
