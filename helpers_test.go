package main

import (
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"path/filepath"

	"time"

	"golang.org/x/net/html"
)

var templates = []string{"default", "chicago"}

var pageToQtype = map[string]string{
	"welcome-page": qTypePage,
	"alpha-single": qTypeOpenSingle,
	"multi-coded":  qTypeCategory,
	"number":       qTypeNumber,
	"open-multi":   qTypeOpenMulti,
	"single-coded": qTypeCategory,
}

func stringForBothTemplates(t *testing.T, page string, test func(string)) {
	for _, template := range templates {
		fileName := filepath.Join("pages", template, page+".html")

		html, err := getHTMLString(fileName)
		assert.NoError(t, err)

		t.Logf("Testing filename '%s'", fileName)
		test(html)
	}
}

func stringForAllQuestionTypes(t *testing.T, test func(string, string)) {
	for page, questionType := range pageToQtype {
		stringForBothTemplates(t, page, func(node string) {
			test(node, questionType)
		})
	}
}

func forBothTemplates(t *testing.T, page string, test func(*html.Node)) {
	stringForBothTemplates(t, page, func(html string) {
		doc, err := htmlStringToNode(html)

		assert.NoError(t, err)

		test(doc)
	})
}

func forAllQuestionTypes(t *testing.T, test func(*html.Node, string)) {
	stringForAllQuestionTypes(t, func(html string, questionType string) {
		doc, err := htmlStringToNode(html)

		assert.NoError(t, err)

		test(doc, questionType)
	})
}

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

func handleRequest(t *testing.T, path string, numberOfRequests *int) pageContent {
	(*numberOfRequests)++

	fileName := filepath.Join(path, fmt.Sprintf("page%d.html", *numberOfRequests))
	url := path

	if isLastFile(path, numberOfRequests) {
		url = endOfInterviewPath
	}

	t.Logf("Doing request for %s", fileName)
	bytes, err := ioutil.ReadFile(fileName)

	if err != nil {
		return pageContent{err: err}
	}

	content := string(bytes)

	return pageContent{body: &content, url: &url}
}

func isLastFile(path string, number *int) bool {
	fileName := filepath.Join(path, fmt.Sprintf("page%d.html", *number+1))
	if _, err := os.Stat(fileName); os.IsNotExist(err) {
		return true
	}

	return false
}

func setupMocking(t *testing.T, path string, numberOfRequests *int) {
	globalConfig = &globalConfiguration{
		verboseOutput:  false,
		requestTimeout: time.Duration(30) * time.Second,
	}
	completeConfig = &completeConfiguration{
		maxConcurrency:   1,
		waitBetweenPosts: time.Duration(0),
		target:           1,
		interviewURL:     path,
	}

	postContent = func(url *string, body url.Values, ch chan pageContent) {
		ch <- handleRequest(t, *url, numberOfRequests)
	}

	getContent = func(url *string, ch chan pageContent) {
		ch <- handleRequest(t, *url, numberOfRequests)
	}
}
