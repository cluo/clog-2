package main

import (
	"path"
	"time"

	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/analysis/lang/en"
	"github.com/blevesearch/bleve/mapping"
)

type ircMsg struct {
	Dt   time.Time `json:"dt"`
	Text string    `json:"text"`
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

	highlightRequest = bleve.NewHighlight()
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

func search(channel, query string) *bleve.SearchResult {
	searchReq := bleve.NewSearchRequest(bleve.NewQueryStringQuery(query))
	searchReq.Highlight = highlightRequest
	index := getIndex(channel)
	results, err := index.Search(searchReq)
	if err != nil {
		panic(err)
	}
	return results
}
