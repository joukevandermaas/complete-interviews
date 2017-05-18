package main

import (
	"fmt"
	"math/rand"
	"net/url"
	"regexp"
	"strings"

	"golang.org/x/net/html"
)

type nodeHandler func(*html.Node)

/**

refactor idea:

scan through the document one time, figure out question type(s)

then using this knowledge, scan through it again to get the correct info

**/

func getInterviewResponse(document *string, previousHistoryOrder string) (url.Values, string, error) {
	writeVerbose(fmt.Sprintf("=============================\n\n%s\n\n=============================\n", *document))
	doc, err := html.Parse(strings.NewReader(*document))

	if err != nil {
		return nil, "", err
	}

	result := url.Values{}
	err = setCommonValues(doc, result)

	if err != nil {
		return nil, "", err
	}

	historyOrder := result["historyOrder"][0]

	if historyOrder == previousHistoryOrder {
		return nil, "", fmt.Errorf("validation error in interview (answer rejected)")
	}

	// assume category question for now
	err = setCategoryQuestionValues(doc, result)

	if err != nil {
		return nil, "", err
	}

	writeVerbose(fmt.Sprintf("posting response: %v\n", result))

	return result, historyOrder, nil
}

func setCommonValues(document *html.Node, result url.Values) error {
	result.Set("button-next", "Next")

	walkDocumentByTag(document, "input", func(input *html.Node) {
		attrs := attrsToMap(input.Attr)

		if attrs["id"] == "screenId" {
			result.Set("screenId", attrs["value"])
		}
		if attrs["id"] == "historyOrder" {
			result.Set("historyOrder", attrs["value"])
		}
	})

	return nil
}

func setCategoryQuestionValues(document *html.Node, result url.Values) error {
	questionRegex, err := regexp.Compile("categorylist-(q\\d+)-multi")
	if err != nil {
		return err
	}

	answerRegex, err := regexp.Compile("answer-(q\\d+)(-\\d+)?")
	if err != nil {
		return err
	}

	var questionNumber string
	var answerOptions []string
	var answerFullValue []string

	walkDocumentByTag(document, "input", func(input *html.Node) {
		attrs := attrsToMap(input.Attr)

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
	})

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

	return nil
}

func walkDocumentByTag(node *html.Node, tag string, handler nodeHandler) {
	walkDocument(node, func(theNode *html.Node) {
		if theNode.Data == tag {
			handler(theNode)
		}
	})
}

func walkDocument(node *html.Node, handler nodeHandler) {
	handler(node)

	if node.FirstChild != nil {
		walkDocument(node.FirstChild, handler)
	}

	if node.NextSibling != nil {
		walkDocument(node.NextSibling, handler)
	}
}

func attrsToMap(attrs []html.Attribute) map[string]string {
	result := make(map[string]string)

	for _, attr := range attrs {
		result[attr.Key] = attr.Val
	}

	return result
}
