package main

import (
	"os"

	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	maxConcurrency   = kingpin.Flag("concurrency", "Maximum number of concurrent interviews (default 10)").Default("10").Int()
	waitBetweenPosts = kingpin.Flag("waitTime", "Wait time in seconds between answering questsions (default 0)").Default("0").Int()

	interviewURL = kingpin.Arg("url", "The url to the interview to complete.").Required().String()
	count        = kingpin.Arg("count", "The number of completes to generate (default 10).").Default("10").Int()
)

func main() {
	kingpin.Parse()

	if !processInterviews() {
		os.Exit(1)
	}
}
