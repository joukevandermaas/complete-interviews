package main

import (
	"bytes"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type pageContent struct {
	body *string
	url  *string

	err error
}

func processInterviews(done *int, errors *int) {
	if *maxConcurrency < 1 {
		*maxConcurrency = 1
	}

	if *maxConcurrency > *count {
		*maxConcurrency = *count
	}

	chInterviews := make(chan *string, *count)
	chResults := make(chan error)

	for i := 0; i < *count; i++ {
		chInterviews <- interviewURL
	}

	for i := 0; i < *maxConcurrency; i++ {
		go func(in chan *string, out chan error) {
			for len(in) > 0 {
				nextInterview := <-in

				err := performInterview(nextInterview)
				out <- err
			}
		}(chInterviews, chResults)
	}

	go writeProgress(done, errors, count)

	for *done < *count {
		err := <-chResults
		(*done)++

		if err != nil {
			(*errors)++
		}
	}
}

func performInterview(url *string) error {
	chInterviews := make(chan pageContent)

	go getContent(url, chInterviews)
	result := <-chInterviews

	if result.err != nil {
		return result.err
	}

	prevHistoryOrder := ""
	hasAnotherQuestion := !strings.Contains(*result.url, endOfInterviewPath)

	for hasAnotherQuestion {
		newRequest, historyOrder, err := getInterviewResponse(result.body, prevHistoryOrder)

		if err != nil {
			return err
		}

		go postContent(result.url, newRequest, chInterviews)

		result = <-chInterviews

		if result.err != nil {
			return err
		}

		hasAnotherQuestion = !strings.Contains(*result.url, endOfInterviewPath)
		prevHistoryOrder = historyOrder
	}

	return nil
}

/* mockable */
var postContent = func(url *string, body url.Values, ch chan pageContent) {
	if *waitBetweenPosts > 0 {
		time.Sleep(time.Duration(*waitBetweenPosts) * time.Second)
	}

	client := http.Client{
		Timeout: requestTimeout,
	}

	response, err := client.Post(*url, "application/x-www-form-urlencoded", strings.NewReader(body.Encode()))

	ch <- handleHTTPResult(response, err)
}

/* mockable */
var getContent = func(url *string, ch chan pageContent) {
	client := http.Client{
		Timeout: requestTimeout,
	}

	response, err := client.Get(*url)

	ch <- handleHTTPResult(response, err)
}

func handleHTTPResult(response *http.Response, err error) pageContent {
	if err != nil {
		return pageContent{err: err}
	}

	defer response.Body.Close()

	url := response.Request.URL.String()

	buf := new(bytes.Buffer)
	buf.ReadFrom(response.Body)

	str := buf.String()

	result := pageContent{body: &str, url: &url}

	if response.StatusCode >= 400 {
		result.err = fmt.Errorf("http request was unsuccessful: %s (url: %s)", response.Status, url)
	}

	return result
}
