package main

import (
	"fmt"
	"math/rand"
	"net/url"
	"regexp"
	"strings"

	"golang.org/x/net/html"
)

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
