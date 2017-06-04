package main

import (
	"math/rand"
	"net/url"
	"os"
	"time"

	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

/* COMMAND LINE OPTIONS */
var (
	requestTimeoutFlag = kingpin.Flag("request-timeout", "Timeout on requests").Default("30s").Duration()
	verboseOutputFlag  = kingpin.Flag("verbose", "Enable verbose output for debugging purposes").Short('v').Default("false").Bool()

	completeCommand              = kingpin.Command("complete", "Complete interviews based on a link").Default()
	completeMaxConcurrencyFlag   = completeCommand.Flag("concurrency", "Maximum number of concurrent interviews").Short('c').Default("10").Int()
	completeWaitBetweenPostsFlag = completeCommand.Flag("wait-time", "Wait time between answering questions").Default("0").Duration()
	completeReplayFileFlag       = completeCommand.Flag("replay-file", "Replay file to determine responses").Short('r').String()
	completeTargetArg            = completeCommand.Arg("count", "The number of completes to generate.").Required().Int()
	completeInterviewURLArg      = completeCommand.Arg("url", "The url to the interview to complete.").Required().String()

	recordCommand         = kingpin.Command("record", "Record an interview for later playback")
	recordOutputFileFlag  = recordCommand.Flag("replay-file", "Output file to write the recording to").Short('r').Default("interview.replay").String()
	recordInterviewURLArg = recordCommand.Arg("url", "The url to the interview to complete.").Required().String()
)

/* GLOBAL DATA STRUCTS */
type completeStatus struct {
	completed int
	errored   int

	lastLinesWritten int
	replaySteps      *[]url.Values
}

type globalConfiguration struct {
	verboseOutput  bool
	requestTimeout time.Duration
	command        string
}

type completeConfiguration struct {
	maxConcurrency   int
	waitBetweenPosts time.Duration
	replayFile       *os.File

	target       int
	interviewURL string
}

type recordConfiguration struct {
	interviewURL string
	replayFile   *os.File
}

var currentStatus *completeStatus
var completeConfig *completeConfiguration
var recordConfig *recordConfiguration
var globalConfig *globalConfiguration

/* STUFF WE NEED */
var random = rand.New(rand.NewSource(time.Now().UnixNano()))
var errorChannel = make(chan error, 100)

const endOfInterviewPath = "/Home/Completed"
