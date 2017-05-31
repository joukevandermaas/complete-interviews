package main

import (
	"math/rand"
	"os"
	"os/signal"
	"strings"
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
var progressBarSize = 30

func main() {
	kingpin.CommandLine.Version("1.0.0")
	kingpin.CommandLine.Help =
		"This tool can complete questionnaires of any number of questions, that " +
			"constist of category, open or number questions, with simple validations " +
			"and no blocks or matrix."
	kingpin.CommandLine.HelpFlag.Short('h')
	kingpin.Parse()

	if *htmlOutputDir != "" {
		writeOutput("Writing html output to '%s/page{n}.html'.\n", *htmlOutputDir)
	}

	done := 0
	errors := 0

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		writeOutput("Stopped completing interviews.%s\n", strings.Repeat(" ", progressBarSize*2))
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
	interviewWord := "interview"
	if *count > 1 {
		interviewWord += "s"
	}
	successful := done - errors

	writeOutput("Finished: successfully completed %d of %d %s (%d%%)%s\n",
		successful,
		*count,
		interviewWord,
		(successful*100) / *count,
		strings.Repeat(" ", progressBarSize))

	if errors > 0 {
		fmt.Fprintf(os.Stderr, "There were %d errors.\n", errors)
	}
}

/* mockable things */
var writeVerbose = func(label string, format string, args ...interface{}) {
	if *verboseOutput {
		fmt.Printf("\r[VERBOSE] "+label+": "+format, args...)
	}
}

var writeOutput = func(format string, args ...interface{}) {
	fmt.Printf("\r"+format, args...)
}

var writeError = func(err error) {
	fmt.Fprintf(os.Stderr, "\rERROR: %v%s\n", err, strings.Repeat(" ", progressBarSize))
}

var writeProgress = func(done *int, errors *int, count *int) {
	spinner := `/-\|/---/|\-`
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
	}
}
