package main

import (
	"os"
	"os/signal"

	"gopkg.in/alecthomas/kingpin.v2"
)

func main() {
	kingpin.CommandLine.Version("1.0.0")
	kingpin.CommandLine.Help =
		"This tool can complete questionnaires of any number of questions, that " +
			"constist of category, open or number questions, with simple validations " +
			"and no blocks or matrix."
	kingpin.CommandLine.HelpFlag.Short('h')
	kingpin.Parse()

	currentStatus = &status{
		completed: 0,
		errored:   0,
	}

	config = &configuration{
		interviewURL: *interviewURLArg,
		target:       *targetArg,

		requestTimeout:   *requestTimeoutFlag,
		waitBetweenPosts: *waitBetweenPostsFlag,
		verboseOutput:    *verboseOutputFlag,
		maxConcurrency:   *maxConcurrencyFlag,
	}

	ensureConsistentOptions()

	printFirstMessage()

	go startOutputLoop()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c

		clearScreen()
		printFinalMessage("Interrupted.")

		os.Exit(1)
	}()

	processInterviews()

	clearScreen()
	printFinalMessage("Finished.")

	if currentStatus.errored > 0 {
		os.Exit(1)
	}
}

func ensureConsistentOptions() {
	if config.target < 1 {
		config.target = 1
	}
	if config.maxConcurrency < 1 {
		config.maxConcurrency = 1
	}
	if config.maxConcurrency > config.target {
		config.maxConcurrency = config.target
	}

	// If stdout is redirected, we want verbose
	// output because the other output is useless
	fi, _ := os.Stdout.Stat()
	if fi != nil {
		config.verboseOutput = true
	}
}
