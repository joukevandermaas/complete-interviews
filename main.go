package main

import (
	"math/rand"
	"os"
	"os/signal"
	"time"

	"fmt"

	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	verboseOutput    = kingpin.Flag("verbose", "Enable verbose output for debugging purposes").Short('v').Default("false").Bool()
	maxConcurrency   = kingpin.Flag("concurrency", "Maximum number of concurrent interviews").Short('c').Default("10").Int()
	waitBetweenPosts = kingpin.Flag("wait-time", "Wait time in seconds between answering questions").Short('w').Default("0").Int()
	htmlOutputDir    = kingpin.Flag("output-html-to", "Enable writing fetched html pages to the specified directory").ExistingDir()

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

	if *htmlOutputDir != "" {
		fmt.Printf("Writing html output to '%s/page{n}.html'.\n", *htmlOutputDir)
	}

	done := 0
	errors := 0

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		writeLastMessage(done, errors)

		os.Exit(1)
	}()

	processInterviews(&done, &errors)

	writeLastMessage(done, errors)

	if errors > 0 {
		os.Exit(1)
	}
}

func writeLastMessage(done int, errors int) {
	fmt.Printf("\nFinished: successfully completed %d of %d interviews\n", done-errors, *count)

	if errors > 0 {
		fmt.Fprintf(os.Stderr, "There were %d errors.\n", errors)
	}
}

/* mockable */
var writeVerbose = func(label string, format string, args ...interface{}) {
	if *verboseOutput {
		fmt.Printf("VERBOSE: "+label+":"+format, args...)
	}
}

/* mockable */
var writeOutput = func(format string, args ...interface{}) {
	fmt.Printf(format, args...)
}

/* mockable */
var writeError = func(err error) {
	fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
}
