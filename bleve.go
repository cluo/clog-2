package main

import (
	"path"

	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/mapping"
)

var (
	indices      map[string]bleve.Index
	indexMapping *mapping.IndexMapping
)

func init() {
	indices = make(map[string]bleve.Index)
}

func getIndex(channel string) bleve.Index {
	index, ok := indices[channel]
	if !ok {
		var err error
		blevePath := path.Join("bleve", channel)
		index, err = bleve.Open(blevePath)
		if err != nil {
			if err.Error() == "cannot open index, path does not exist" {
				mapping := bleve.NewIndexMapping()
				index, err = bleve.New(blevePath, mapping)
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
	searchReq := bleve.NewSearchRequest(bleve.NewMatchQuery(query))
	index := getIndex(channel)
	results, err := index.Search(searchReq)
	if err != nil {
		panic(err)
	}
	return results
}
