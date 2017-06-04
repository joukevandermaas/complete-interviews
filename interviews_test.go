package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCompleteInterviews(t *testing.T) {
	numberOfRequests := 0
	setupMocking(t, "pages/test-interview", &numberOfRequests)

	assert := assert.New(t)

	err := performInterview(&completeConfig.interviewURL)
	assert.NoError(err)

	assert.Equal(13, numberOfRequests)
}
