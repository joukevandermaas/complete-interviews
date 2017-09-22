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

	"golang.org/x/net/html"
)

type pageContent struct {
	body *string
	url  *string

	err error
}

type interviewToComplete struct {
	url    *string
	number int
}

func processInterviews() {
	chInterviews := make(chan interviewToComplete, completeConfig.target)
	chResults := make(chan error, completeConfig.target)

	go func() {
		var replaySteps []url.Values
		if completeConfig.replayFile != nil {
			replaySteps = parseReplayFile(completeConfig.replayFile)
			currentStatus.replaySteps = &replaySteps
		}

		for i := 0; i < completeConfig.target; i++ {
			chInterviews <- interviewToComplete{url: &completeConfig.interviewURL, number: i}
		}

		for i := 0; i < completeConfig.maxConcurrency; i++ {
			if i > 0 {
				if completeConfig.waitBetweenPosts > 0 {
					time.Sleep(completeConfig.waitBetweenPosts)
				} else {
					time.Sleep(50 * time.Millisecond)
				}
			}
			go func(in chan interviewToComplete, out chan error) {
				printVerbose("thread", "Starting thread...\n")

				for len(in) > 0 {
					nextInterview := <-in

					var err error
					if globalConfig.command == "complete" {
						err = performInterview(nextInterview.url, nextInterview.number)
					} else if globalConfig.command == "replay" {
						err = performReplay(nextInterview.url)
					} else {
						err = fmt.Errorf("WTF are you doing")
					}

					out <- err
				}

				printVerbose("thread", "Thread finished.\n")
			}(chInterviews, chResults)
		}
	}()

	go func() {
		for currentStatus.completed < completeConfig.target {
			time.Sleep(500 * time.Millisecond)

			currentStatus.active = completeConfig.target - len(chInterviews) - currentStatus.completed
		}
	}()

	for currentStatus.completed < completeConfig.target {
		err := <-chResults
		currentStatus.completed++

		if err != nil {
			printError(err)
			currentStatus.errored++
		}
	}

	currentStatus.active = 0
}

func performReplay(url *string) error {
	chInterviews := make(chan pageContent)

	go getContent(url, chInterviews)
	result := <-chInterviews

	if result.err != nil {
		return result.err
	}

	findScreenID := func(content pageContent) (string, error) {
		foundScreenID := ""
		doc, err := html.Parse(strings.NewReader(*content.body))

		if err != nil {
			return foundScreenID, err
		}

		walkDocumentByTag(doc, "input", func(input *html.Node) {
			attrs := attrsToMap(input.Attr)

			if attrs["id"] == "screenId" {
				foundScreenID = attrs["value"]
			}
		})

		return foundScreenID, nil
	}

	screenID, err := findScreenID(result)

	if err != nil {
		return err
	}

	for _, answers := range *currentStatus.replaySteps {
		if strings.Contains(*result.url, endOfInterviewPath) {
			// start new interview; replay contained multiple
			printVerbose("replay", "Starting new interview, because replay file is longer.\n")
			go getContent(url, chInterviews)
			result = <-chInterviews

			if result.err != nil {
				return result.err
			}

			screenID, err = findScreenID(result)

			if err != nil {
				return err
			}
		}

		response := addScreenID(answers, screenID)
		printVerbose("replay", "posting %v\n", response)
		go postContent(result.url, answers, chInterviews)
		result = <-chInterviews

		if result.err != nil {
			return result.err
		}
	}

	if !strings.Contains(*result.url, endOfInterviewPath) {
		return fmt.Errorf("end of replay file did not result in completed interview")
	}

	return nil
}

func performInterview(url *string, number int) error {
	chInterviews := make(chan pageContent)

	startURL := *url
	if completeConfig.respondentKeyFormat != "" {
		if !strings.HasSuffix(startURL, "/") {
			startURL += "/"
		}
		startURL += fmt.Sprintf(completeConfig.respondentKeyFormat, number)
	}

	go getContent(&startURL, chInterviews)
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
			return result.err
		}

		hasAnotherQuestion = !strings.Contains(*result.url, endOfInterviewPath)
		prevHistoryOrder = historyOrder
	}

	return nil
}

func addScreenID(form url.Values, screenID string) url.Values {
	result := url.Values{}
	result.Set("screenId", screenID)

	for key, values := range form {
		for _, value := range values {
			result.Add(key, value)
		}
	}

	return result
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
	lines := strings.FieldsFunc(question, func(char rune) bool { return char == '\n' })
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
