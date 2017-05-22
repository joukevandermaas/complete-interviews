package main

import (
	"io/ioutil"
	"net/url"
	"strings"

	"golang.org/x/net/html"
)

func htmlStringToNode(snippet string) (*html.Node, error) {
	doc, err := html.Parse(strings.NewReader(snippet))
	if err != nil {
		return nil, err
	}

	return doc, nil
}

func getHTMLString(filename string) (string, error) {
	file, err := ioutil.ReadFile(filename)
	if err != nil {
		return "", err
	}

	return string(file), nil
}

func getHTMLDocument(filename string) (*html.Node, error) {
	content, err := getHTMLString(filename)
	if err != nil {
		return nil, err
	}

	return htmlStringToNode(content)
}

func flattenURLValues(values url.Values) map[string]string {
	result := make(map[string]string)

	for key, val := range values {
		result[key] = strings.Join(val, ",")
	}

	return result
}
