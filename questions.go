package main

import (
	"fmt"
	"math/rand"
	"net/url"
	"regexp"
	"strings"

	lorem "github.com/drhodes/golorem"

	"strconv"

	"golang.org/x/net/html"
)

type nodeHandler func(*html.Node)

const (
	qTypeOpenMulti  = "OpenMulti"
	qTypeOpenSingle = "OpenSingle"
	qTypeCategory   = "Category"
	qTypePage       = "Page"
)

func getInterviewResponse(document *string, previousHistoryOrder string) (url.Values, string, error) {
	doc, err := html.Parse(strings.NewReader(*document))

	if err != nil {
		return nil, "", err
	}

	result := url.Values{}
	err = setCommonValues(doc, result)

	if err != nil {
		return nil, "", err
	}

	writeVerbose("common-values", fmt.Sprintf("%v\n", result))

	var historyOrder string

	if val, ok := result["historyOrder"]; ok {
		historyOrder = val[0]

		if historyOrder == previousHistoryOrder {
			return nil, "", fmt.Errorf("validation error in interview (answer rejected)")
		}
	}

	questionType := getQuestionType(doc)

	writeVerbose("question-type", fmt.Sprintf("%s\n", questionType))

	switch questionType {
	case qTypeCategory:
		err = setCategoryQuestionValues(doc, result)
	case qTypeOpenMulti:
		err = setOpenMultiQuestionValues(doc, result)
	case qTypeOpenSingle:
		err = setOpenSingleQuestionValues(doc, result)
	}

	if err != nil {
		return nil, "", err
	}

	writeVerbose("posting response", fmt.Sprintf("%v\n", result))

	return result, historyOrder, nil
}

func getQuestionType(document *html.Node) string {
	foundTextArea := false
	foundCategoryInput := false
	foundQuestionInput := false

	categoryRegexp := regexp.MustCompile("categorylist-(q\\d+)-multi")
	questionRegexp := regexp.MustCompile("q\\d+")

	walkDocument(document, func(node *html.Node) {
		if node.Data == "textarea" {
			foundTextArea = true
		} else if node.Data == "input" {
			attrs := attrsToMap(node.Attr)
			if categoryRegexp.MatchString(attrs["id"]) {
				foundCategoryInput = true
			}

			hasSimpleID := questionRegexp.MatchString(attrs["id"])
			hasAnswerName := attrs["name"] == "answer-"+attrs["id"]
			typeIsText := attrs["type"] == "text"

			if hasSimpleID && hasAnswerName && typeIsText {

				foundQuestionInput = true
			}
		}
	})

	if foundTextArea {
		return qTypeOpenMulti
	} else if foundCategoryInput {
		return qTypeCategory
	} else if foundQuestionInput {
		return qTypeOpenSingle
	}

	return qTypePage
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

func setOpenMultiQuestionValues(document *html.Node, result url.Values) error {
	walkDocumentByTag(document, "textarea", func(node *html.Node) {
		attrs := attrsToMap(node.Attr)

		result.Set(attrs["name"], lorem.Paragraph(2, 5))
	})

	return nil
}

func setOpenSingleQuestionValues(document *html.Node, result url.Values) error {
	questionRegexp := regexp.MustCompile("q\\d+")
	var innerError error

	walkDocumentByTag(document, "input", func(node *html.Node) {
		attrs := attrsToMap(node.Attr)

		if questionRegexp.MatchString(attrs["id"]) {
			minLength := 0
			maxLength := 250

			if strVal, ok := attrs["minlength"]; ok {
				value, err := parseInt(strVal)
				if err != nil {
					innerError = err
					return
				}
				minLength = value
			}
			if strVal, ok := attrs["maxlength"]; ok {
				value, err := parseInt(strVal)
				if err != nil {
					innerError = err
					return
				}
				maxLength = value
			}

			result.Set(attrs["name"], lorem.Word(minLength, maxLength))
		}
	})

	return innerError
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
		writeVerbose("attrs", fmt.Sprintf("%v\n", attrs))

		if len(matched) > 0 {
			writeVerbose("matched questions", fmt.Sprintf("%v\n", matched[0]))
			questionNumber = matched[0][1]

			return // read: move on to next input
		}

		matched = answerRegex.FindAllStringSubmatch(attrs["name"], 1)

		if len(matched) > 0 {
			writeVerbose("matched answers", fmt.Sprintf("%v\n", matched[0]))
			answerOptions = append(answerOptions, strings.TrimPrefix(attrs["value"], questionNumber+"-"))

			if len(matched[0]) > 2 {
				fullValue := fmt.Sprintf("%s%s", matched[0][1], matched[0][2])
				writeVerbose("full value", fmt.Sprintf("%s (>2)\n", fullValue))
				answerFullValue = append(answerFullValue, fullValue)
			} else {
				fullValue := matched[0][1]
				writeVerbose("full value", fmt.Sprintf("%s (>1)\n", fullValue))
				answerFullValue = append(answerFullValue, fullValue)
			}
		}
	})

	if len(answerOptions) > 0 {
		pickedAnswerIndex := rand.Intn(len(answerOptions))
		pickedAnswer := answerOptions[pickedAnswerIndex]
		pickedAnswerDescription := answerFullValue[pickedAnswerIndex]

		writeVerbose("picked answer", fmt.Sprintf("%d\n", pickedAnswerIndex))
		writeVerbose("options", fmt.Sprintf("%v\n", answerOptions))
		writeVerbose("descriptions", fmt.Sprintf("%v\n", answerFullValue))

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

func parseInt(value string) (int, error) {
	result, err := strconv.ParseInt(value, 0, 32)

	if err != nil {
		return 0, err
	}

	return int(result), nil
}
