package main

import (
	"os"
	"os/signal"

	tm "github.com/buger/goterm"
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

	tm.Clear()

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
	/*interviewWord := "interview"
	if *count > 1 {
		interviewWord += "s"
	}
	successful := done - errors*/
}

/* mockable things */

var writeProgress = func(done *int, errors *int, count *int) {
	/*spinner := `/-\|/---/|\-`
	index := 0
	percPerBlock := 100 / progressBarSize

	for *done < *count {
		percentDone := *done * 100 / *count
		doneBlocks := percentDone / percPerBlock

		if doneBlocks > progressBarSize {
			doneBlocks = progressBarSize
		}

		if *verboseOutput {
			writeOutput("completed: %4d of %4d (%2d%%), %4d errors\n",
				*done,
				*count,
				percentDone,
				*errors)
			time.Sleep(4 * time.Second)
		} else {
			writeOutput("%s%s [%d%%] %s completed: %d of %d, %d errors",
				strings.Repeat("▓", doneBlocks),
				strings.Repeat("░", progressBarSize-doneBlocks),
				percentDone,
				string(spinner[index]),
				*done,
				*count,
				*errors)
			time.Sleep(250 * time.Millisecond)
		}

		index = (index + 1) % len(spinner)
	}*/
}
