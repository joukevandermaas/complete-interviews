package main

import (
	"fmt"
	"net/url"
	"testing"

	"golang.org/x/net/html"

	"strconv"

	"github.com/stretchr/testify/assert"
)

func TestSetOpenMultiQuestionValues(t *testing.T) {
	assert := assert.New(t)

	forBothTemplates(t, "open-multi", func(doc *html.Node) {
		values := make(url.Values)

		err := setOpenMultiQuestionValues(doc, values)
		assert.NoError(err)

		result := flattenURLValues(values)

		t.Logf("%v\n", result)

		assert.NotEmpty(result["answer-q1"])
	})
}

func TestSetCommonValues(t *testing.T) {
	assert := assert.New(t)

	forAllQuestionTypes(t, func(doc *html.Node, questionType string) {
		values := make(url.Values)

		err := setCommonValues(doc, values)
		assert.NoError(err)

		result := flattenURLValues(values)

		t.Logf("%v\n", result)

		assert.Equal("0", result["historyOrder"])
		assert.Equal("032794ea-dfbb-4c33-95c2-2fbe5befd885", result["screenId"])
		assert.Equal("Next", result["button-next"])
	})
}

func TestSetOpenSingleQuestionValues(t *testing.T) {
	assert := assert.New(t)

	forBothTemplates(t, "alpha-single", func(doc *html.Node) {
		values := make(url.Values)

		err := setOpenSingleQuestionValues(doc, values)
		assert.NoError(err)

		result := flattenURLValues(values)

		t.Logf("%v\n", result)

		assert.NotEmpty(result["answer-q1"])
	})
}

func TestSetSingleCategoryValues(t *testing.T) {
	assert := assert.New(t)

	forBothTemplates(t, "single-coded", func(doc *html.Node) {
		values := make(url.Values)

		err := setCategoryQuestionValues(doc, values)
		assert.NoError(err)

		result := flattenURLValues(values)

		t.Logf("%v\n", result)

		answer1, err := strconv.ParseInt(result["answer-q1-m"], 0, 32)
		assert.NoError(err)

		assert.True(answer1 >= 0 && answer1 <= 5, "answer in range 1-5")
		assert.Equal(fmt.Sprintf("q1-%d", answer1), result["answer-q1"])
	})
}

func TestSetMultiCategoryValues(t *testing.T) {
	assert := assert.New(t)

	forBothTemplates(t, "multi-coded", func(doc *html.Node) {
		values := make(url.Values)

		err := setCategoryQuestionValues(doc, values)
		assert.NoError(err)

		answers := values["answer-q1-m"]

		assert.Len(answers, 4)

		for _, answer := range answers {
			answerInt, err := strconv.ParseInt(answer, 0, 32)
			assert.NoError(err)

			category := fmt.Sprintf("q1-%d", answerInt)

			assert.True(answerInt >= 1 && answerInt <= 10)
			assert.Equal(category, values[fmt.Sprintf("answer-%s", category)][0])
		}
	})
}

func TestGetQuestionType(t *testing.T) {
	assert := assert.New(t)

	forAllQuestionTypes(t, func(doc *html.Node, questionType string) {
		result := getQuestionType(doc)

		assert.Equal(questionType, result)
	})
}

func TestGetInterviewResponseReturnsErrOnSameHistoryOrder(t *testing.T) {
	assert := assert.New(t)

	stringForAllQuestionTypes(t, func(doc string, _ string) {
		_, _, err := getInterviewResponse(&doc, "0")

		assert.Error(err)
	})
}

func TestGetInterviewResponseReturnsGoodResponseForMultiLineText(t *testing.T) {
	assert := assert.New(t)

	stringForBothTemplates(t, "open-multi", func(doc string) {
		response, historyOrder, err := getInterviewResponse(&doc, "")
		assert.NoError(err)

		assert.Equal("0", historyOrder)

		result := flattenURLValues(response)
		t.Logf("%v\n", result)

		assert.NotEmpty(result["answer-q1"])
		assert.True(len(result["answer-q1"]) > 10, "Length greater than 10")
	})
}

func TestGetInterviewResponseReturnsGoodResponseForSingleCategory(t *testing.T) {
	assert := assert.New(t)

	stringForBothTemplates(t, "single-coded", func(doc string) {
		response, historyOrder, err := getInterviewResponse(&doc, "")
		assert.NoError(err)

		assert.Equal("0", historyOrder)

		result := flattenURLValues(response)
		t.Logf("%v\n", result)

		answer1, err := strconv.ParseInt(result["answer-q1-m"], 0, 32)
		assert.NoError(err)

		assert.True(answer1 >= 1 && answer1 <= 5, "should be in range 1-5")
		assert.Equal(fmt.Sprintf("q1-%d", answer1), result["answer-q1"])
	})
}

func TestGetInterviewResponseReturnsGoodResponseForMultiCategory(t *testing.T) {
	assert := assert.New(t)

	stringForBothTemplates(t, "multi-coded", func(doc string) {
		response, historyOrder, err := getInterviewResponse(&doc, "")
		assert.NoError(err)

		assert.Equal("0", historyOrder)

		result := response["answer-q1-m"][0]
		t.Logf("%v\n", result)

		answer1, err := strconv.ParseInt(result, 0, 32)
		assert.NoError(err)

		category := fmt.Sprintf("q1-%d", answer1)

		assert.True(answer1 >= 1 && answer1 <= 10, "answer should be in range 1-10")
		assert.Equal(category, response[fmt.Sprintf("answer-%s", category)][0])
	})
}

func TestGetInterviewResponseReturnsGoodResponseForSingleLineText(t *testing.T) {
	assert := assert.New(t)

	stringForBothTemplates(t, "alpha-single", func(doc string) {
		response, historyOrder, err := getInterviewResponse(&doc, "")
		assert.NoError(err)

		assert.Equal("0", historyOrder)

		result := flattenURLValues(response)
		t.Logf("%v\n", result)

		assert.NotEmpty(result["answer-q1"])
		assert.True(len(result["answer-q1"]) < 13, "Length smaller than 13")
	})
}

func TestGetInterviewResponseReturnsGoodResponseForNumber(t *testing.T) {
	assert := assert.New(t)

	stringForBothTemplates(t, "number", func(doc string) {
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
	})
}

func TestGetInterviewResponseReturnsGoodResponseForWelcomePage(t *testing.T) {
	assert := assert.New(t)

	stringForBothTemplates(t, "welcome-page", func(doc string) {
		response, historyOrder, err := getInterviewResponse(&doc, "")
		assert.NoError(err)

		result := flattenURLValues(response)

		assert.Equal("0", historyOrder)
		assert.Empty(result["answer-q1"])
	})
}
