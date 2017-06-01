package main

import (
	"fmt"
	"math"
	"strings"
	"time"

	"os"

	tm "github.com/buger/goterm"
)

func printError(err error) {
	if !config.verboseOutput {
		errorChannel <- err
	} else {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
	}
}

func printVerbose(context string, format string, args ...interface{}) {
	if config.verboseOutput {
		fmt.Printf("["+context+"] "+format, args...)
	}
}

func printFirstMessage() {
	lines := []string{}

	var endOfSentence string
	if config.target > 1 {
		endOfSentence = fmt.Sprintf("s, at most %d concurrently.", config.maxConcurrency)
	} else {
		endOfSentence = "."
	}

	lines = addLine(lines, "Will complete %d interview%s", config.target, endOfSentence)

	if config.waitBetweenPosts > 0 {
		lines = addLine(lines, "Waiting %s between questions.", config.waitBetweenPosts.String())
	}

	flushLines(lines)
}

func printFinalMessage(reason string) {
	lines := []string{}

	addBasicStatusLines(&lines)

	lines = addLine(lines, "")
	lines = addLine(lines, "%s Completed %d of %d interviews.", reason, currentStatus.completed, config.target)

	flushLines(lines)
}

func addBasicStatusLines(lines *[]string) {
	if !config.verboseOutput {
		*lines = addLine(*lines, "Successful : %4d", currentStatus.completed-currentStatus.errored)
		*lines = addLine(*lines, "Error      : %4d", currentStatus.errored)
	} else {
		*lines = addLine(*lines, "Successful: %4d, Error: %4d",
			currentStatus.completed-currentStatus.errored, currentStatus.errored)
	}
}

func startOutputLoop() {
	spinner := `/-\|`
	frameIndex := 0

	lines := []string{}
	for currentStatus.completed < config.target {
		select {
		case err := <-errorChannel:
			// some error happened
			emptyLine := strings.Repeat(" ", tm.Width())
			tm.Printf("ERROR: %v\n%s\n", err, emptyLine)
		default:
			// no errors yet
		}

		s := currentStatus
		percentDone := s.completed * 100 / config.target

		if !config.verboseOutput {
			statusLine := fmt.Sprintf("[%s] %d of %d interviews (%d%%)", string(spinner[frameIndex]), s.completed, config.target, percentDone)
			progressBar := getProgressBar(tm.Width() - len(statusLine) - 1)

			lines = addLine(lines, "%s %s", statusLine, progressBar)
			lines = addLine(lines, "")
		}

		addBasicStatusLines(&lines)

		flushLines(lines)

		if !config.verboseOutput {
			tm.MoveCursorUp(len(lines) + 1)
			s.lastLinesWritten = len(lines)
			frameIndex = (frameIndex + 1) % len(spinner)
			time.Sleep(250 * time.Millisecond)
		} else {
			time.Sleep(4 * time.Second)
		}

		lines = lines[:0]
	}
}

func getProgressBar(size int) string {
	s := currentStatus

	fraction := float64(s.completed) / float64(config.target)
	doneBlocks := int(math.Ceil(fraction * float64(size)))

	return strings.Repeat("▓", doneBlocks) + strings.Repeat("░", size-doneBlocks)
}

func addLine(lines []string, format string, arguments ...interface{}) []string {
	return append(lines, fmt.Sprintf(format, arguments...))
}

func flushLines(lines []string) {
	for _, line := range lines {
		if !config.verboseOutput {
			tm.Println(line)
		} else {
			fmt.Println(line)
		}
	}

	if !config.verboseOutput {
		tm.Flush()
	}
}

func clearScreen() {
	if !config.verboseOutput {
		lines := []string{}
		for i := 0; i < currentStatus.lastLinesWritten; i++ {
			lines = append(lines, strings.Repeat(" ", tm.Width()))
		}
		flushLines(lines)
		tm.MoveCursorUp(len(lines) + 1)
	}
}
