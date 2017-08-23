package main

import (
	"fmt"
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

	command := kingpin.Parse()

	globalConfig = &globalConfiguration{
		requestTimeout: *requestTimeoutFlag,
		verboseOutput:  *verboseOutputFlag,
		command:        command,
	}

	// If stdout is redirected, we want verbose
	// output because the other output is useless
	fi, _ := os.Stdout.Stat()
	if fi != nil {
		globalConfig.verboseOutput = true
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		clearScreen()
		printFinalMessage("Interrupted.")

		os.Exit(1)
	}()

	switch command {
	case "complete":
		executeCompleteCommand()
	case "record":
		executeRecordCommand()
	case "replay":
		executeReplayCommand()
	}
}

func executeRecordCommand() {
	file, err := os.Create(*recordOutputFileFlag)
	defer file.Close()

	recordConfig = &recordConfiguration{
		interviewURL: *recordInterviewURLArg,
		target:       *recordTargetArg,
		replayFile:   file,
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	startProxyForInterview()

	printFinalMessage("Done.")
}

func executeCompleteCommand() {
	currentStatus = &completeStatus{
		completed: 0,
		errored:   0,
	}

	completeConfig = &completeConfiguration{
		interviewURL: *completeInterviewURLArg,
		target:       *completeTargetArg,

		waitBetweenPosts:    *completeWaitBetweenPostsFlag,
		maxConcurrency:      *completeMaxConcurrencyFlag,
		respondentKeyFormat: getGolangFormat(*completeRespondentKeyFormatFlag),
	}

	ensureConsistentCompleteOptions()
	printFirstMessage()

	go startOutputLoop()

	processInterviews()

	clearScreen()
	printFinalMessage("Finished.")

	if currentStatus.errored > 0 {
		os.Exit(1)
	}
}

func executeReplayCommand() {
	currentStatus = &completeStatus{
		completed: 0,
		errored:   0,
	}

	completeConfig = &completeConfiguration{
		interviewURL: *replayInterviewURLArg,
		target:       *replayTargetArg,

		waitBetweenPosts:    *replayWaitBetweenPostsFlag,
		maxConcurrency:      *replayMaxConcurrencyFlag,
		respondentKeyFormat: "",
	}

	file, err := os.Open(*replayFileArg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
	completeConfig.replayFile = file

	ensureConsistentCompleteOptions()
	printFirstMessage()

	go startOutputLoop()

	processInterviews()

	clearScreen()
	printFinalMessage("Finished.")

	if currentStatus.errored > 0 {
		os.Exit(1)
	}
}

func ensureConsistentCompleteOptions() {
	if completeConfig.target < 1 {
		completeConfig.target = 1
	}
	if completeConfig.maxConcurrency < 1 {
		completeConfig.maxConcurrency = 1
	}
	if completeConfig.maxConcurrency > completeConfig.target {
		completeConfig.maxConcurrency = completeConfig.target
	}
}

func getGolangFormat(cmdLineFormat string) string {
	resultFormat := ""
	previousCharacter := '\x00'
	percentCount := 0
	numberFormat := "%%0%dd"

	for _, character := range cmdLineFormat {
		if character == '%' {
			if percentCount == 0 || previousCharacter == '%' {
				percentCount++
			} else {
				kingpin.FatalUsage("Invalid format string '%s'. Percentage signs must not be separated by anything.", cmdLineFormat)
			}
		} else {
			if previousCharacter == '%' {
				resultFormat += fmt.Sprintf(numberFormat, percentCount)
			}
			resultFormat += string(character)
		}

		previousCharacter = character
	}

	if previousCharacter == '%' {
		resultFormat += fmt.Sprintf(numberFormat, percentCount)
	}

	if percentCount == 0 {
		kingpin.FatalUsage("Invalid format string '%s'. Must have at least one percentage sign.", cmdLineFormat)
	}

	return resultFormat
}
