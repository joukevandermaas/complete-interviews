package main

import (
	"bytes"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
)

type pageContent struct {
	body string
	url  string
}

func main() {
	url := "https://rc-interviewing.niposoftware-dev.com/Interviews/wwuaw/NKtRYwUmedoyizz48yiU"

	chInterviews := make(chan pageContent)

	go getContent(url, chInterviews)
	result := <-chInterviews

	questionNumber := 0
	hasAnotherQuestion := true // depends on result.url

	for hasAnotherQuestion {
		newRequest := getInterviewResponse(result.body, questionNumber)
		go postContent(result.url, newRequest, chInterviews)
		result = <-chInterviews

		questionNumber++
		hasAnotherQuestion = false // depends on result.url
	}

	os.Exit(0)
}

func getInterviewResponse(html string, questionNumber int) url.Values {
	fmt.Print(html)
	// parse html, find question & possible answers and form result
	return nil
}

func postContent(url string, body url.Values, ch chan pageContent) {
	response, err := http.Post(url, "application/x-www-form-urlencoded", strings.NewReader(body.Encode()))
	if err != nil {
		ch <- pageContent{}
	} else {
		defer response.Body.Close()

		url := response.Request.URL.String()

		buf := new(bytes.Buffer)
		buf.ReadFrom(response.Body)
		str := buf.String()

		result := pageContent{body: str, url: url}
		ch <- result
	}
}

func getContent(url string, ch chan pageContent) {
	response, err := http.Get(url)
	if err != nil {
		ch <- pageContent{}
	} else {
		defer response.Body.Close()

		url := response.Request.URL.String()

		buf := new(bytes.Buffer)
		buf.ReadFrom(response.Body)
		str := buf.String()

		result := pageContent{body: str, url: url}
		ch <- result
	}
}
