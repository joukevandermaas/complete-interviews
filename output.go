package main

import (
	"fmt"
	"math"
	"strings"
	"time"

	tm "github.com/buger/goterm"
)

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
	lines = addLine(lines, reason)

	flushLines(lines)
}

func addBasicStatusLines(lines *[]string) {
	*lines = addLine(*lines, "Successful : %4d", currentStatus.completed-currentStatus.errored)
	*lines = addLine(*lines, "Error      : %4d", currentStatus.errored)
}

func printProgress() {
	spinner := `/-\|`
	frameIndex := 0

	lines := []string{}
	for currentStatus.completed < config.target {
		s := currentStatus
		percentDone := s.completed * 100 / config.target

		statusLine := fmt.Sprintf("[%s] %d of %d interviews (%d%%)", string(spinner[frameIndex]), s.completed, config.target, percentDone)
		progressBar := getProgressBar(tm.Width() - len(statusLine) - 1)

		lines = addLine(lines, "%s %s", statusLine, progressBar)
		lines = addLine(lines, "")

		addBasicStatusLines(&lines)

		flushLines(lines)
		tm.MoveCursorUp(len(lines) + 1)
		s.lastLinesWritten = len(lines)

		time.Sleep(250 * time.Millisecond)
		frameIndex = (frameIndex + 1) % len(spinner)

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
		tm.Println(line)
	}

	tm.Flush()
}

func clearScreen() {
	lines := []string{}
	for i := 0; i < currentStatus.lastLinesWritten; i++ {
		lines = append(lines, strings.Repeat(" ", tm.Width()))
	}
	flushLines(lines)
	tm.MoveCursorUp(len(lines) + 1)
}
