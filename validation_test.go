package main

import (
	"net/url"
	"testing"

	"strings"

	"strconv"

	"github.com/stretchr/testify/assert"
)

func TestNumberValidations(t *testing.T) {
	assert := assert.New(t)

	// note: multiple answers required
	doc, err := htmlStringToNode(`
<input id="q1" type="text" class="open number required" value="" name="answer-q1" data-minimum="1" data-maximum="3" />
`)
	assert.NoError(err)

	values := make(url.Values)

	err = setNumberQuestionValues(doc, values)
	assert.NoError(err)

	result := flattenURLValues(values)

	t.Logf("%v\n", result)

	answer, err := strconv.ParseInt(result["answer-q1"], 0, 32)
	assert.NoError(err)

	assert.True(answer >= 1, "At least 1")
	assert.True(answer <= 3, "At most 3")
}

func TestMultiCategoryValidations(t *testing.T) {
	assert := assert.New(t)

	// note: multiple answers required
	doc, err := htmlStringToNode(`
<div id="categorylist-q1" data-minimum="2" data-maximum="4">
	<input type="hidden" class="answerOrder" name="answer-q1-m" id="categorylist-q1-multi" value="" />
	<div class="categorygroup"">
		<input id="q1-1" class="category" name="answer-q1-1" value="q1-1" type="checkbox" />
		<input id="q1-2" class="category" name="answer-q1-2" value="q1-2" type="checkbox" />
		<input id="q1-3" class="category" name="answer-q1-3" value="q1-3" type="checkbox" />
		<input id="q1-4" class="category" name="answer-q1-4" value="q1-4" type="checkbox" />
		<input id="q1-5" class="category" name="answer-q1-5" value="q1-5" type="checkbox" />
		<input id="q1-6" class="category" name="answer-q1-6" value="q1-6" type="checkbox" />
	</div>
</div>
`)
	assert.NoError(err)

	values := make(url.Values)

	err = setCategoryQuestionValues(doc, values)
	assert.NoError(err)

	result := flattenURLValues(values)

	t.Logf("%v\n", result)

	answerMulti := strings.Split(result["answer-q1-m"], ",")
	t.Logf("%v\n", answerMulti)

	assert.True(len(answerMulti) >= 2, "At least 2 answers given")
	assert.True(len(answerMulti) <= 4, "At most 4 answers given")

	for _, answer := range answerMulti {
		assert.Equal("q1-"+answer, result["answer-q1-"+answer])
	}
}
