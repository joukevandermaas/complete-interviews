package main

import (
	"fmt"
	"net/url"
	"testing"

	"strconv"

	"github.com/stretchr/testify/assert"
)

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

func TestSetOpenSingleQuestionValues(t *testing.T) {
	assert := assert.New(t)

	doc, err := getHTMLDocument("test-files/singleline.html")
	assert.NoError(err)

	values := make(url.Values)

	err = setOpenSingleQuestionValues(doc, values)
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

	values := map[string]string{
		"test-files/multi-category.html":  qTypeCategory,
		"test-files/single-category.html": qTypeCategory,
		"test-files/multiline.html":       qTypeOpenMulti,
		"test-files/number.html":          qTypeNumber,
		"test-files/singleline.html":      qTypeOpenSingle,
		"test-files/welcome-page.html":    qTypePage,
		"test-files/page-question.html":   qTypePage,
	}

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

func TestGetInterviewResponseReturnsGoodResponseForMultiLineText(t *testing.T) {
	assert := assert.New(t)

	doc, err := getHTMLString("test-files/multiline.html")
	assert.NoError(err)

	response, historyOrder, err := getInterviewResponse(&doc, "")
	assert.NoError(err)

	assert.Equal("0", historyOrder)

	result := flattenURLValues(response)
	t.Logf("%v\n", result)

	assert.NotEmpty(result["answer-q1"])
	assert.True(len(result["answer-q1"]) > 10, "Length greater than 10")
}

func TestGetInterviewResponseReturnsGoodResponseForSingleCategory(t *testing.T) {
	assert := assert.New(t)

	doc, err := getHTMLString("test-files/single-category.html")
	assert.NoError(err)

	response, historyOrder, err := getInterviewResponse(&doc, "")
	assert.NoError(err)

	assert.Equal("0", historyOrder)

	result := flattenURLValues(response)
	t.Logf("%v\n", result)

	answer1, err := strconv.ParseInt(result["answer-q1-m"], 0, 32)
	assert.NoError(err)

	assert.True(answer1 >= 0 && answer1 <= 3)
	assert.Equal(fmt.Sprintf("q1-%d", answer1), result["answer-q1"])
}

func TestGetInterviewResponseReturnsGoodResponseForMultiCategory(t *testing.T) {
	assert := assert.New(t)

	doc, err := getHTMLString("test-files/multi-category.html")
	assert.NoError(err)

	response, historyOrder, err := getInterviewResponse(&doc, "")
	assert.NoError(err)

	assert.Equal("0", historyOrder)

	result := flattenURLValues(response)
	t.Logf("%v\n", result)

	answer1, err := strconv.ParseInt(result["answer-q1-m"], 0, 32)
	assert.NoError(err)

	category := fmt.Sprintf("q1-%d", answer1)

	assert.True(answer1 >= 0 && answer1 <= 3)
	assert.Equal(category, result[fmt.Sprintf("answer-%s", category)])
}

func TestGetInterviewResponseReturnsGoodResponseForSingleLineText(t *testing.T) {
	assert := assert.New(t)

	doc, err := getHTMLString("test-files/singleline.html")
	assert.NoError(err)

	response, historyOrder, err := getInterviewResponse(&doc, "")
	assert.NoError(err)

	assert.Equal("0", historyOrder)

	result := flattenURLValues(response)
	t.Logf("%v\n", result)

	assert.NotEmpty(result["answer-q1"])
	assert.True(len(result["answer-q1"]) > 4, "Length greater than 4")
	assert.True(len(result["answer-q1"]) < 13, "Length smaller than 13")
}

func TestGetInterviewResponseReturnsGoodResponseForNumber(t *testing.T) {
	assert := assert.New(t)

	doc, err := getHTMLString("test-files/number.html")
	assert.NoError(err)

	response, historyOrder, err := getInterviewResponse(&doc, "")
	assert.NoError(err)

	assert.Equal("0", historyOrder)

	result := flattenURLValues(response)
	t.Logf("%v\n", result)

	assert.NotEmpty(result["answer-q1"])

	value, err := strconv.ParseInt(result["answer-q1"], 0, 32)
	assert.NoError(err)

	assert.True(value >= 5, "Value greater than 5")
	assert.True(value <= 15, "Value smaller than 15")
}

func TestGetInterviewResponseReturnsGoodResponseForWelcomePage(t *testing.T) {
	assert := assert.New(t)

	doc, err := getHTMLString("test-files/welcome-page.html")
	assert.NoError(err)

	response, historyOrder, err := getInterviewResponse(&doc, "")
	assert.NoError(err)

	result := flattenURLValues(response)

	assert.Equal("0", historyOrder)
	assert.Empty(result["answer-q1"])
}

func TestGetInterviewResponseReturnsGoodResponseForPageQuestion(t *testing.T) {
	assert := assert.New(t)

	doc, err := getHTMLString("test-files/page-question.html")
	assert.NoError(err)

	response, historyOrder, err := getInterviewResponse(&doc, "")
	assert.NoError(err)

	result := flattenURLValues(response)

	assert.Equal("0", historyOrder)
	assert.Empty(result["answer-q1"])
}
