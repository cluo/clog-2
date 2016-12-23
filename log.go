package main

import (
	"compress/gzip"
	"fmt"
	"io"
	"os"
)

type logfile struct {
	reader    io.ReadCloser
	rawReader io.ReadCloser
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
	return &logfile{reader: reader, rawReader: rawReader}, nil
}

func (lf *logfile) Close() {
	lf.reader.Close()
	if lf.rawReader != nil {
		lf.rawReader.Close()
	}
}

type logNotFound struct {
	relpath string
}

func (lnf *logNotFound) Error() string {
	return fmt.Sprintf("couldn't find \"%s\"", lnf.relpath)
}
