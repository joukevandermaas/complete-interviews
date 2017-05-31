package main

import (
	"math/rand"
	"time"

	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

type status struct {
	completed int
	errored   int

	lastLinesWritten int
}

type configuration struct {
	verboseOutput    bool
	maxConcurrency   int
	waitBetweenPosts time.Duration
	requestTimeout   time.Duration

	target       int
	interviewURL string
}

var currentStatus *status
var config *configuration

var (
	verboseOutputFlag    = kingpin.Flag("verbose", "Enable verbose output for debugging purposes").Short('v').Default("false").Bool()
	maxConcurrencyFlag   = kingpin.Flag("concurrency", "Maximum number of concurrent interviews").Short('c').Default("10").Int()
	waitBetweenPostsFlag = kingpin.Flag("wait-time", "Wait time between answering questions").Default("0").Duration()
	requestTimeoutFlag   = kingpin.Flag("request-timeout", "Timeout on requests").Default("30s").Duration()

	targetArg       = kingpin.Arg("count", "The number of completes to generate.").Required().Int()
	interviewURLArg = kingpin.Arg("url", "The url to the interview to complete.").Required().String()
)

var random = rand.New(rand.NewSource(time.Now().UnixNano()))

const endOfInterviewPath = "/Home/Completed"
