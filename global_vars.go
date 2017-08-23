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

	completeCommand                 = kingpin.Command("complete", "Complete interviews based on a link").Default()
	completeMaxConcurrencyFlag      = completeCommand.Flag("concurrency", "Maximum number of concurrent interviews").Short('c').Default("10").Int()
	completeWaitBetweenPostsFlag    = completeCommand.Flag("wait-time", "Wait time between answering questions").Default("0").Duration()
	completeRespondentKeyFormatFlag = completeCommand.Flag("respondent-key", "Format for respondent key").Default("").String()
	completeTargetArg               = completeCommand.Arg("count", "The number of completes to generate.").Required().Int()
	completeInterviewURLArg         = completeCommand.Arg("url", "The url to the interview to complete.").Required().String()

	recordCommand         = kingpin.Command("record", "Record an interview for later playback")
	recordOutputFileFlag  = recordCommand.Flag("replay-file", "Output file to write the recording to").Short('r').Default("interview.replay").String()
	recordTargetArg       = recordCommand.Arg("count", "The number of completes to record.").Required().Int()
	recordInterviewURLArg = recordCommand.Arg("url", "The url to the interview to complete.").Required().String()

	replayCommand              = kingpin.Command("replay", "Replay interviews based on a replay file")
	replayMaxConcurrencyFlag   = replayCommand.Flag("concurrency", "Maximum number of concurrent interviews").Short('c').Default("10").Int()
	replayWaitBetweenPostsFlag = replayCommand.Flag("wait-time", "Wait time between answering questions").Default("0").Duration()
	replayTargetArg            = replayCommand.Arg("count", "The number of replays to generate.").Required().Int()
	replayInterviewURLArg      = replayCommand.Arg("url", "The url to the interview to complete.").Required().String()
	replayFileArg              = replayCommand.Arg("replay-file", "Replay file to determine responses").Default("interview.replay").String()
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

	respondentKeyFormat string
}

type recordConfiguration struct {
	target       int
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
