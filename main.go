package main

import (
	"math/rand"
	"os"
	"time"

	"fmt"

	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	verboseOutput    = kingpin.Flag("verbose", "Enable verbose output for debugging purposes").Short('v').Default("false").Bool()
	maxConcurrency   = kingpin.Flag("concurrency", "Maximum number of concurrent interviews").Short('c').Default("10").Int()
	waitBetweenPosts = kingpin.Flag("wait-time", "Wait time in seconds between answering questions").Short('w').Default("0").Int()

	count        = kingpin.Arg("count", "The number of completes to generate.").Required().Int()
	interviewURL = kingpin.Arg("url", "The url to the interview to complete.").Required().String()
)

var random = rand.New(rand.NewSource(time.Now().UnixNano()))

func main() {
	kingpin.CommandLine.Version("1.0.0")
	kingpin.CommandLine.Help =
		"This tool can complete questionnaires of any number of questions, that " +
			"constist of category, open or number questions, with simple validations " +
			"and no blocks or matrix."
	kingpin.CommandLine.HelpFlag.Short('h')
	kingpin.Parse()

	errorCount := processInterviews()

	if errorCount > 0 {
		fmt.Fprintf(os.Stderr, "There were %d errors.\n", errorCount)
		os.Exit(1)
	}
}

func writeVerbose(label string, message string) {
	if *verboseOutput {
		fmt.Printf("VERBOSE: %s: %s", label, message)
	}
}
