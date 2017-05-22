package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type pageContent struct {
	body *string
	url  *string

	err error
}

var requestTimeout = time.Duration(30 * time.Second)
var endOfInterviewPath = "/Home/Completed"

func processInterviews() int {
	requestURL := (*interviewURL).String()

	if *maxConcurrency < 1 {
		*maxConcurrency = 1
	}

	if *maxConcurrency > *count {
		*maxConcurrency = *count
	}

	chInterviews := make(chan *string, *count)
	chResults := make(chan error)

	machine := func(in chan *string, out chan error) {
		resultChannel := make(chan error)
		for len(in) > 0 {
			nextInterview := <-in

			writeVerbose("picked up interview", "%s\n", *nextInterview)

			go performInterview(nextInterview, resultChannel)
			out <- <-resultChannel
		}

		writeVerbose("thread", "Stopping thread to process interviews...\n")
	}

	var waitTimeString string
	if *waitBetweenPosts > 0 {
		waitTimeString = fmt.Sprintf(", waiting %ds between questions", *waitBetweenPosts)
	} else {
		waitTimeString = ""
	}

	fmt.Printf("Starting interviews (%d concurrently%s)...\n", *maxConcurrency, waitTimeString)
	for i := 0; i < *count; i++ {
		chInterviews <- &requestURL
	}

	for i := 0; i < *maxConcurrency; i++ {
		writeVerbose("thread", "Starting thread to process interviews...\n")
		go machine(chInterviews, chResults)
	}

	var printNo int

	if *count > 1000 {
		printNo = 100
	} else if *count > 100 {
		printNo = 50
	} else if *count > 10 {
		printNo = 10
	} else {
		printNo = 1
	}

	errors := 0
	done := 0
	for done < *count {
		if done%printNo == 0 {
			fmt.Printf("completed: %4d of %d\n", done, *count)
		}

		err := <-chResults
		done++

		writeVerbose("info", "done: %4d; errors: %4d; queue: %4d\n", done, errors, len(chInterviews))

		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
			errors++
		}
	}

	fmt.Printf("\nFinished: successfully completed %d of %d interviews\n", done-errors, *count)

	return errors
}

func performInterview(url *string, ch chan error) {
	chInterviews := make(chan pageContent)

	go getContent(url, chInterviews)
	result := <-chInterviews

	if result.err != nil {
		ch <- result.err

		return
	}

	prevHistoryOrder := ""
	hasAnotherQuestion := !strings.Contains(*result.url, endOfInterviewPath)

	for hasAnotherQuestion {
		newRequest, historyOrder, err := getInterviewResponse(result.body, prevHistoryOrder)

		if err != nil {
			ch <- err
		}

		go postContent(result.url, newRequest, chInterviews)

		result = <-chInterviews

		if result.err != nil {
			ch <- result.err
			return
		}

		hasAnotherQuestion = !strings.Contains(*result.url, endOfInterviewPath)
		prevHistoryOrder = historyOrder

		writeVerbose("================================", "\n")
	}

	ch <- nil
}

func postContent(url *string, body url.Values, ch chan pageContent) {
	if *waitBetweenPosts > 0 {
		time.Sleep(time.Duration(*waitBetweenPosts) * time.Second)
	}

	client := http.Client{
		Timeout: requestTimeout,
	}

	response, err := client.Post(*url, "application/x-www-form-urlencoded", strings.NewReader(body.Encode()))

	ch <- handleHTTPResult(response, err)
}

func getContent(url *string, ch chan pageContent) {
	client := http.Client{
		Timeout: requestTimeout,
	}

	response, err := client.Get(*url)

	ch <- handleHTTPResult(response, err)
}

var httpTime = 1

func handleHTTPResult(response *http.Response, err error) pageContent {
	if err != nil {
		return pageContent{err: err}
	}

	defer response.Body.Close()

	url := response.Request.URL.String()

	buf := new(bytes.Buffer)
	buf.ReadFrom(response.Body)

	if *htmlOutputDir != "" {
		bytes := buf.Bytes()
		err = ioutil.WriteFile(filepath.Join(*htmlOutputDir, fmt.Sprintf("page%d.html", httpTime)), bytes, os.ModeAppend)

		if err != nil {
			return pageContent{err: err}
		}
	}

	str := buf.String()

	result := pageContent{body: &str, url: &url}

	if response.StatusCode >= 400 {
		result.err = fmt.Errorf("http request was unsuccessful: %s (url: %s)", response.Status, url)
	}

	httpTime++
	return result
}
