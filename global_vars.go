package main

import (
	"math/rand"
	"time"

	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

type progress struct {
	target     int
	successful int
	errored    int
}

var (
	verboseOutput    = kingpin.Flag("verbose", "Enable verbose output for debugging purposes").Short('v').Default("false").Bool()
	maxConcurrency   = kingpin.Flag("concurrency", "Maximum number of concurrent interviews").Short('c').Default("10").Int()
	waitBetweenPosts = kingpin.Flag("wait-time", "Wait time in seconds between answering questions").Short('w').Default("0").Int()

	count        = kingpin.Arg("count", "The number of completes to generate.").Required().Int()
	interviewURL = kingpin.Arg("url", "The url to the interview to complete.").Required().String()
)

var random = rand.New(rand.NewSource(time.Now().UnixNano()))

var requestTimeout = time.Duration(30 * time.Second)
var endOfInterviewPath = "/Home/Completed"
