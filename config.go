package main

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type configuration struct {
	Irclogs    string
	Channels   []string
	RereadHtml bool
}

var config configuration

func getConfig() configuration {
	var c configuration
	contents, err := ioutil.ReadFile("config.yaml")
	if err != nil {
		panic(err)
	} else {
		yaml.Unmarshal(contents, &c)
		return c
	}
}
