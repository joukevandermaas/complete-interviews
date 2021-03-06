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
	if !globalConfig.verboseOutput && isCompletingInterviews() {
		errorChannel <- err
	} else {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
	}
}

func printVerbose(context string, format string, args ...interface{}) {
	if globalConfig.verboseOutput {
		fmt.Printf("VERBOSE: ["+context+"] "+format, args...)
	}
}

func printFirstMessage() {
	if isCompletingInterviews() {
		lines := []string{}

		whatAreWeDoing := "interview"

		if globalConfig.command == "replay" {
			whatAreWeDoing = "replay playthrough"
		}

		var endOfSentence string
		if completeConfig.target > 1 {
			endOfSentence = fmt.Sprintf("s, at most %d concurrently.", completeConfig.maxConcurrency)
		} else {
			endOfSentence = "."
		}

		lines = addLine(lines, "Will complete %d %s%s", completeConfig.target, whatAreWeDoing, endOfSentence)

		if globalConfig.command == "replay" {
			lines = addLine(lines, "Using replay file \"%s\"", completeConfig.replayFile.Name())
		}

		if completeConfig.waitBetweenPosts > 0 {
			lines = addLine(lines, "Waiting %s between questions.", completeConfig.waitBetweenPosts.String())
		}

		flushLines(lines)
	}
}

func printFinalMessage(reason string) {
	if isCompletingInterviews() {
		lines := []string{}

		for len(errorChannel) > 0 {
			err := <-errorChannel
			width := tm.Width()
			emptyLine := strings.Repeat(" ", width)

			line := fmt.Sprintf("%v", err)
			length := len([]rune(line))
			if length < width {
				line += strings.Repeat(" ", width-length)
			}
			tm.Printf("ERROR: %s\n%s\n", line, emptyLine)
		}

		addBasicStatusLines(&lines)

		whatAreWeDoing := "interviews"

		if globalConfig.command == "replay" {
			whatAreWeDoing = "replay playthroughs"
		}

		lines = addLine(lines, strings.Repeat(" ", tm.Width()))
		lines = addLine(lines, "%s Completed %d of %d %s.", reason, currentStatus.completed, completeConfig.target, whatAreWeDoing)

		flushLines(lines)
	} else if globalConfig.command == "record" {
		fmt.Printf("%s\n", reason)
	}
}

func addBasicStatusLines(lines *[]string) {
	if !globalConfig.verboseOutput {
		*lines = addLine(*lines, "Successful : %4d", currentStatus.completed-currentStatus.errored)
		*lines = addLine(*lines, "Error      : %4d", currentStatus.errored)

		if currentStatus.active > 0 {
			*lines = addLine(*lines, "Active     : %4d", currentStatus.active)
		}
	} else {
		*lines = addLine(*lines, "Successful: %4d, Error: %4d",
			currentStatus.completed-currentStatus.errored, currentStatus.errored)
	}
}

func startOutputLoop() {
	spinner := []rune(`⠁⠁⠉⠙⠚⠒⠂⠂⠒⠲⠴⠤⠄⠄⠤⠠⠠⠤⠦⠖⠒⠐⠐⠒⠓⠋⠉⠈⠈`)
	frameIndex := 0

	lines := []string{}
	for currentStatus.completed < completeConfig.target {
		select {
		case err := <-errorChannel:
			// some error happened
			emptyLine := strings.Repeat(" ", tm.Width())
			tm.Printf("ERROR: %v\n%s\n", err, emptyLine)
		default:
			// no errors yet
		}

		s := currentStatus
		percentDone := s.completed * 100 / completeConfig.target

		if !globalConfig.verboseOutput {
			whatAreWeDoing := "interviews"

			if globalConfig.command == "replay" {
				whatAreWeDoing = "replay playthroughs"
			}

			statusLine := fmt.Sprintf("%d of %d %s (%d%%)",
				s.completed,
				completeConfig.target,
				whatAreWeDoing,
				percentDone)
			progressBar := getProgressBar(tm.Width() - 1)

			lines = addLine(lines, "[%s] %s", string(spinner[frameIndex]), statusLine)
			lines = addLine(lines, "")
			lines = addLine(lines, progressBar)
			lines = addLine(lines, strings.Repeat(" ", tm.Width()))
		}

		addBasicStatusLines(&lines)

		flushLines(lines)

		if !globalConfig.verboseOutput {
			tm.MoveCursorUp(len(lines) + 1)
			s.lastLinesWritten = len(lines)
			frameIndex = (frameIndex + 1) % len(spinner)
			time.Sleep(50 * time.Millisecond)
		} else {
			time.Sleep(4 * time.Second)
		}

		lines = lines[:0]
	}
}

func getProgressBar(size int) string {
	s := currentStatus

	fraction := float64(s.completed) / float64(completeConfig.target)
	doneBlocks := int(math.Ceil(fraction * float64(size)))

	return strings.Repeat("▓", doneBlocks) + strings.Repeat("░", size-doneBlocks)
}

func addLine(lines []string, format string, arguments ...interface{}) []string {
	line := fmt.Sprintf(format, arguments...)

	length := len([]rune(line))

	if length < tm.Width() {
		line += strings.Repeat(" ", tm.Width()-length)
	}

	return append(lines, line)
}

func flushLines(lines []string) {
	for _, line := range lines {
		if !globalConfig.verboseOutput {
			tm.Println(line)
		} else {
			fmt.Println(line)
		}
	}

	if !globalConfig.verboseOutput {
		tm.Flush()
	}
}

func clearScreen() {
	if !globalConfig.verboseOutput && isCompletingInterviews() {
		lines := []string{}
		for i := 0; i < currentStatus.lastLinesWritten; i++ {
			lines = append(lines, strings.Repeat(" ", tm.Width()))
		}
		flushLines(lines)
		tm.MoveCursorUp(len(lines) + 1)
	}
}

func isCompletingInterviews() bool {
	return globalConfig.command == "complete" || globalConfig.command == "replay"
}
