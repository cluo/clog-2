package main

import (
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/analysis/lang/en"
	"github.com/blevesearch/bleve/mapping"
	"github.com/blevesearch/bleve/search/highlight/format/ansi"
	_ "github.com/blevesearch/bleve/search/highlight/highlighter/ansi"
)

type ircMsg struct {
	Dt   time.Time `json:"dt"`
	Text string    `json:"text"`
}

type searchResult struct {
	Date string `json:"date"`
	Line int    `json:"line"`
	Text string `json:"text"`
}

var (
	indices          map[string]bleve.Index
	indexMapping     *mapping.IndexMappingImpl
	highlightRequest *bleve.HighlightRequest
)

func init() {
	indices = make(map[string]bleve.Index)

	indexMapping = bleve.NewIndexMapping()
	indexMapping.DefaultMapping.AddFieldMappingsAt("dt", bleve.NewDateTimeFieldMapping())
	tfm := bleve.NewTextFieldMapping()
	tfm.Analyzer = en.AnalyzerName
	indexMapping.DefaultMapping.AddFieldMappingsAt("text", tfm)

	highlightRequest = bleve.NewHighlightWithStyle("ansi")
	highlightRequest.AddField("text")
}

func getIndex(channel string) bleve.Index {
	index, ok := indices[channel]
	if !ok {
		var err error
		blevePath := path.Join("bleve", channel)
		index, err = bleve.Open(blevePath)
		if err != nil {
			if err == bleve.ErrorIndexPathDoesNotExist {
				index, err = bleve.New(blevePath, indexMapping)
				if err != nil {
					panic(err)
				}
			} else {
				panic(err)
			}
		}
		indices[channel] = index
	}
	return index
}

func search(channel, query string) ([]searchResult, error) {
	searchReq := bleve.NewSearchRequest(bleve.NewQueryStringQuery(query))
	searchReq.Highlight = highlightRequest

	index := getIndex(channel)
	results, err := index.Search(searchReq)
	if err != nil {
		return nil, err
	}
	hits := make([]searchResult, len(results.Hits))
	for i, hit := range results.Hits {
		split := strings.SplitN(hit.ID, ":", 2)
		hits[i].Date = split[0]
		fragment := hit.Fragments["text"][0]
		fragmentLines := strings.Split(fragment, "\n")
		var cleanedFragment string
		for _, fragmentLine := range fragmentLines {
			if strings.Index(fragmentLine, ansi.DefaultAnsiHighlight) >= 0 {
				cleanedFragment = strings.Replace(fragmentLine, ansi.DefaultAnsiHighlight, "", -1)
				cleanedFragment = strings.Replace(cleanedFragment, ansi.Reset, "", -1)
				cleanedFragment = strings.Replace(cleanedFragment, "…", "", -1)
				break
			}
		}
		sixHourWindow, err := strconv.Atoi(split[1])
		if err != nil {
			return nil, err
		}
		lf, err := openLog(path.Join(channel, hits[i].Date))
		if err != nil {
			return nil, err
		}
		defer lf.Close()
		hits[i].Line, hits[i].Text = lf.Line(sixHourWindow, cleanedFragment)
	}
	return hits, nil
}
