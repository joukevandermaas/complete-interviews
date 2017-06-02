package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

type pageContent struct {
	body *string
	url  *string

	err error
}

func processInterviews() {
	chInterviews := make(chan *string, completeConfig.target)
	chResults := make(chan error)

	var replaySteps []url.Values
	if completeConfig.replayFile != nil {
		replaySteps = parseReplayFile(completeConfig.replayFile)
		currentStatus.replaySteps = &replaySteps
	}

	for i := 0; i < completeConfig.target; i++ {
		chInterviews <- &completeConfig.interviewURL
	}

	for i := 0; i < completeConfig.maxConcurrency; i++ {
		go func(in chan *string, out chan error) {
			printVerbose("thread", "Starting thread...\n")

			for len(in) > 0 {
				nextInterview := <-in

				err := performInterview(nextInterview)
				out <- err
			}

			printVerbose("thread", "Thread finished.\n")
		}(chInterviews, chResults)
	}

	for currentStatus.completed < completeConfig.target {
		err := <-chResults
		currentStatus.completed++

		if err != nil {
			printError(err)
			currentStatus.errored++
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

	if currentStatus.replaySteps == nil {
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
				return result.err
			}

			hasAnotherQuestion = !strings.Contains(*result.url, endOfInterviewPath)
			prevHistoryOrder = historyOrder
		}
	} else {
		for _, answers := range *currentStatus.replaySteps {
			printVerbose("replay", "posting %v\n", answers)
			go postContent(result.url, answers, chInterviews)
			result = <-chInterviews

			if result.err != nil {
				return result.err
			}
		}
		if !strings.Contains(*result.url, endOfInterviewPath) {
			return fmt.Errorf("end of replay file did not result in completed interview")
		}
	}

	return nil
}

func parseReplayFile(file *os.File) []url.Values {
	buf := bytes.NewBuffer(nil)
	io.Copy(buf, file) // Error handling elided for brevity.

	questions := strings.Split(string(buf.Bytes()), "---\n")
	answers := []url.Values{}

	for _, question := range questions {
		if strings.TrimSpace(question) != "" {
			answers = append(answers, parseReplayQuestion(question))
		}
	}

	return answers
}

func parseReplayQuestion(question string) url.Values {
	lines := strings.Fields(question)
	printVerbose("replay", "question\n")
	result := url.Values{}

	for _, line := range lines {
		splitLine := strings.Split(line, "=")
		key := splitLine[0]
		valuesString := splitLine[1]

		values := strings.Trim(valuesString, "[]")

		printVerbose("replay", "key: %s, value: %s\n", key, values)

		result.Set(key, values)
	}

	return result
}

/* mockable */
var postContent = func(url *string, body url.Values, ch chan pageContent) {
	if completeConfig.waitBetweenPosts > 0 {
		time.Sleep(completeConfig.waitBetweenPosts)
	}

	client := http.Client{
		Timeout: globalConfig.requestTimeout,
	}

	response, err := client.Post(*url, "application/x-www-form-urlencoded", strings.NewReader(body.Encode()))

	ch <- handleHTTPResult(response, err)
}

/* mockable */
var getContent = func(url *string, ch chan pageContent) {
	client := http.Client{
		Timeout: globalConfig.requestTimeout,
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
