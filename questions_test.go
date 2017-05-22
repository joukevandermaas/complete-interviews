package main

import (
	"fmt"
	"io/ioutil"
	"net/url"
	"strings"
	"testing"

	"golang.org/x/net/html"

	"strconv"

	"github.com/stretchr/testify/assert"
)

func TestSetCommonValues(t *testing.T) {
	assert := assert.New(t)

	doc, err := getHTMLDocument("test-files/welcome-page.html")
	assert.NoError(err)
	values := make(url.Values)

	err = setCommonValues(doc, values)
	assert.NoError(err)

	result := flattenURLValues(values)

	t.Logf("%v\n", result)

	assert.Equal("0", result["historyOrder"])
	assert.Equal("032794ea-dfbb-4c33-95c2-2fbe5befd885", result["screenId"])
	assert.Equal("Next", result["button-next"])
}

func TestSetOpenMultiQuestionValues(t *testing.T) {
	assert := assert.New(t)

	doc, err := getHTMLDocument("test-files/multiline.html")
	assert.NoError(err)

	values := make(url.Values)

	err = setOpenMultiQuestionValues(doc, values)
	assert.NoError(err)

	result := flattenURLValues(values)

	t.Logf("%v\n", result)

	assert.NotEmpty(result["answer-q1"])
}

func TestSetSingleCategoryValues(t *testing.T) {
	assert := assert.New(t)

	doc, err := getHTMLDocument("test-files/single-category.html")
	assert.NoError(err)

	values := make(url.Values)

	err = setCategoryQuestionValues(doc, values)
	assert.NoError(err)

	result := flattenURLValues(values)

	t.Logf("%v\n", result)

	answer1, err := strconv.ParseInt(result["answer-q1-m"], 0, 32)
	assert.NoError(err)

	assert.True(answer1 >= 0 && answer1 <= 3)
	assert.Equal(fmt.Sprintf("q1-%d", answer1), result["answer-q1"])
}

func TestSetMultiCategoryValues(t *testing.T) {
	assert := assert.New(t)

	doc, err := getHTMLDocument("test-files/multi-category.html")
	assert.NoError(err)

	values := make(url.Values)

	err = setCategoryQuestionValues(doc, values)
	assert.NoError(err)

	result := flattenURLValues(values)

	t.Logf("%v\n", result)

	answer1, err := strconv.ParseInt(result["answer-q1-m"], 0, 32)
	assert.NoError(err)

	category := fmt.Sprintf("q1-%d", answer1)

	assert.True(answer1 >= 0 && answer1 <= 3)
	assert.Equal(category, result[fmt.Sprintf("answer-%s", category)])
}

func TestGetQuestionType(t *testing.T) {
	assert := assert.New(t)

	values := make(map[string]string)

	values["test-files/multi-category.html"] = qTypeCategory
	values["test-files/single-category.html"] = qTypeCategory
	values["test-files/multiline.html"] = qTypeOpenMulti
	values["test-files/welcome-page.html"] = qTypePage
	values["test-files/page-question.html"] = qTypePage

	for file, questionType := range values {
		doc, err := getHTMLDocument(file)
		assert.NoError(err)

		result := getQuestionType(doc)

		assert.Equal(questionType, result, file)
	}
}

func TestGetInterviewResponseReturnsErrOnSameHistoryOrder(t *testing.T) {
	assert := assert.New(t)

	doc, err := getHTMLString("test-files/single-category.html")
	assert.NoError(err)

	_, _, err = getInterviewResponse(&doc, "0")

	assert.Error(err)
}

/********************************************************************************************\
|                                     test helper functions                                  |
\********************************************************************************************/

func getHTMLString(filename string) (string, error) {
	file, err := ioutil.ReadFile(filename)
	if err != nil {
		return "", err
	}

	return string(file), nil
}

func getHTMLDocument(filename string) (*html.Node, error) {
	content, err := getHTMLString(filename)
	if err != nil {
		return nil, err
	}

	doc, err := html.Parse(strings.NewReader(content))
	if err != nil {
		return nil, err
	}

	return doc, nil
}

func flattenURLValues(values url.Values) map[string]string {
	result := make(map[string]string)

	for key, val := range values {
		result[key] = val[0]
	}

	return result
}
