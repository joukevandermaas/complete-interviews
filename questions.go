package main

import (
	"fmt"
	"math/rand"
	"net/url"
	"regexp"
	"strings"

	"golang.org/x/net/html"
)

type nodeHandler func(*html.Node) error

func getInterviewResponse(document *string, previousHistoryOrder string) (url.Values, string, error) {
	writeVerbose(fmt.Sprintf("=============================\n\n%s\n\n=============================\n", *document))
	doc, err := html.Parse(strings.NewReader(*document))

	if err != nil {
		return nil, "", err
	}

	result := url.Values{}
	result.Set("button-next", "Next")

	questionRegex, err := regexp.Compile("categorylist-(q\\d+)-multi")
	if err != nil {
		return nil, "", err
	}

	answerRegex, err := regexp.Compile("answer-(q\\d+)(-\\d+)?")
	if err != nil {
		return nil, "", err
	}

	var questionNumber string
	var answerOptions []string
	var answerFullValue []string
	var historyOrder string

	err = walkDocument(doc, "input", func(input *html.Node) error {
		attrs := attrsToMap(input.Attr)

		if attrs["id"] == "screenId" {
			result.Set("screenId", attrs["value"])
		}
		if attrs["id"] == "historyOrder" {
			historyOrder = attrs["value"]
			if previousHistoryOrder == historyOrder {
				return fmt.Errorf("validation error in interview (answer rejected)")
			}
			result.Set("historyOrder", historyOrder)
		}

		matched := questionRegex.FindAllStringSubmatch(attrs["id"], 1)

		writeVerbose(fmt.Sprintf("%v\n", matched))

		if len(matched) > 0 {
			questionNumber = matched[0][1]
		}

		matched = answerRegex.FindAllStringSubmatch(attrs["name"], 1)

		if len(matched) > 0 {
			answerOptions = append(answerOptions, strings.TrimPrefix(attrs["value"], questionNumber+"-"))

			if len(matched[0]) > 2 {
				answerFullValue = append(answerFullValue, fmt.Sprintf("%s%s", matched[0][1], matched[0][2]))
			} else {
				answerFullValue = append(answerFullValue, matched[0][1])
			}
		}

		return nil
	})

	if err != nil {
		return nil, "", err
	}

	if len(answerOptions) > 0 {
		pickedAnswerIndex := rand.Intn(len(answerOptions))
		pickedAnswer := answerOptions[pickedAnswerIndex]
		pickedAnswerDescription := answerFullValue[pickedAnswerIndex]

		result.Set(
			fmt.Sprintf("answer-%s-m", questionNumber),
			pickedAnswer)
		result.Set(
			fmt.Sprintf("answer-%s", pickedAnswerDescription),
			fmt.Sprintf("%s-%s", questionNumber, pickedAnswer))
	}

	writeVerbose(fmt.Sprintf("posting response: %v\n", result))

	return result, historyOrder, nil
}

func walkDocument(node *html.Node, tag string, handler nodeHandler) error {
	if node.Data == tag {
		err := handler(node)
		if err != nil {
			return err
		}
	}

	if node.FirstChild != nil {
		err := walkDocument(node.FirstChild, tag, handler)
		if err != nil {
			return err
		}
	}

	if node.NextSibling != nil {
		err := walkDocument(node.NextSibling, tag, handler)
		if err != nil {
			return err
		}
	}

	return nil
}

func attrsToMap(attrs []html.Attribute) map[string]string {
	result := make(map[string]string)

	for _, attr := range attrs {
		result[attr.Key] = attr.Val
	}

	return result
}
