package main

import (
	"bytes"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"golang.org/x/net/html"
)

type pageContent struct {
	body *string
	url  *string

	err error
}

type nodeHandler func(*html.Node)

func processInterviews() bool {
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
		for {
			nextInterview := <-in
			go performInterview(nextInterview, resultChannel)
			out <- <-resultChannel
		}
	}

	var waitTimeString string
	if *waitBetweenPosts > 0 {
		waitTimeString = fmt.Sprintf(", waiting %ds between questions", *waitBetweenPosts)
	} else {
		waitTimeString = ""
	}

	fmt.Printf("Starting interviews (%d concurrently%s)...\n", *maxConcurrency, waitTimeString)
	for i := 0; i < *count; i++ {
		chInterviews <- interviewURL
	}

	for i := 0; i < *maxConcurrency; i++ {
		go machine(chInterviews, chResults)
	}

	errors := 0
	done := 0
	for done < *count {
		if done%*maxConcurrency == 0 {
			fmt.Printf("completed %d of %d interviews\n", done, *count)
		}

		err := <-chResults
		done++

		if err != nil {
			fmt.Printf("ERROR: %v\n", err)
			errors++
		}
	}

	fmt.Printf("\nFinished: successfully completed %d of %d interviews\n", done-errors, *count)

	return errors > 0
}

func performInterview(url *string, ch chan error) {
	chInterviews := make(chan pageContent)

	go getContent(url, chInterviews)
	result := <-chInterviews

	if result.err != nil {
		ch <- result.err
	}

	hasAnotherQuestion := !strings.Contains(*result.url, "/Home/Completed")

	for hasAnotherQuestion {
		newRequest, err := getInterviewResponse(result.body)

		if err != nil {
			ch <- err
		}

		go postContent(result.url, newRequest, chInterviews)

		result = <-chInterviews

		if result.err != nil {
			ch <- result.err
		}

		hasAnotherQuestion = !strings.Contains(*result.url, "/Home/Completed")
	}

	ch <- nil
}

func getInterviewResponse(document *string) (url.Values, error) {
	doc, err := html.Parse(strings.NewReader(*document))

	if err != nil {
		return nil, err
	}

	result := url.Values{}
	result.Set("button-next", "Next")

	questionRegex, err := regexp.Compile("categorylist-(q\\d+)-multi")

	if err != nil {
		return nil, err
	}

	var questionNumber string
	var answerOptions []string

	walkDocument(doc, "input", func(input *html.Node) {
		attrs := attrsToMap(input.Attr)

		if attrs["id"] == "screenId" {
			result.Set("screenId", attrs["value"])
		}
		if attrs["id"] == "historyOrder" {
			result.Set("historyOrder", attrs["value"])
		}

		matched := questionRegex.FindAllStringSubmatch(attrs["id"], 1)

		if len(matched) > 0 {
			questionNumber = matched[0][1]
		}

		if attrs["name"] == "answer-"+questionNumber {
			answerOptions = append(answerOptions, strings.TrimPrefix(attrs["value"], questionNumber+"-"))
		}

	})

	if len(answerOptions) > 0 {
		pickedAnswer := answerOptions[rand.Intn(len(answerOptions))]

		result.Set(
			fmt.Sprintf("answer-%s-m", questionNumber),
			pickedAnswer)
		result.Set(
			fmt.Sprintf("answer-%s", questionNumber),
			fmt.Sprintf("%s-%s", questionNumber, pickedAnswer))
	}

	return result, nil
}

func walkDocument(node *html.Node, tag string, handler nodeHandler) {
	if node.Data == tag {
		handler(node)
	}

	if node.FirstChild != nil {
		walkDocument(node.FirstChild, tag, handler)
	}

	if node.NextSibling != nil {
		walkDocument(node.NextSibling, tag, handler)
	}
}

func attrsToMap(attrs []html.Attribute) map[string]string {
	result := make(map[string]string)

	for _, attr := range attrs {
		result[attr.Key] = attr.Val
	}

	return result
}

func postContent(url *string, body url.Values, ch chan pageContent) {

	if *waitBetweenPosts > 0 {
		time.Sleep(time.Duration(*waitBetweenPosts) * time.Second)
	}

	response, err := http.Post(*url, "application/x-www-form-urlencoded", strings.NewReader(body.Encode()))

	ch <- handleHTTPResult(response, err)
}

func getContent(url *string, ch chan pageContent) {
	response, err := http.Get(*url)

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
