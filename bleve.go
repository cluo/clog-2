package main

import (
	"path"
	"time"

	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/mapping"
)

type ircMsg struct {
	Dt time.Time
	Text string
}

var (
	indices      map[string]bleve.Index
	indexMapping *mapping.IndexMappingImpl
)

func init() {
	indices = make(map[string]bleve.Index)
	indexMapping = bleve.NewIndexMapping()
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
	index := getIndex(channel)
	results, err := index.Search(searchReq)
	if err != nil {
		panic(err)
	}
	return results
}
