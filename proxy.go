package main

import (
	"bytes"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"time"
)

func startProxyForInterview() {
	requests := 0
	isDone := false

	handleRequest := func(request *http.Request) {
		if request.Method == "POST" {
			request.ParseForm()
			form := request.Form

			printVerbose("recording", "Recording interview answer %v\n", form)
			writeResponseToFile(recordConfig.replayFile, form)
		}
	}

	redirectAtEndOfInterview := func(response http.ResponseWriter, request *http.Request) {
		if strings.Contains(request.URL.String(), endOfInterviewPath) {
			if requests < (recordConfig.target - 1) {
				requests++
				fmt.Printf("Completed interview %d of %d\n", requests, recordConfig.target)
				http.Redirect(response, request, "/", 302)
			} else {
				isDone = true
			}
		}
	}

	isLastRequest := func(url string) bool {
		willStop := strings.Contains(url, endOfInterviewPath) && isDone

		if willStop {
			printVerbose("recording", "Last interview is done, stopping server now.\n")
		}
		return willStop
	}

	runProxy(recordConfig.interviewURL, handleRequest, redirectAtEndOfInterview, isLastRequest)

	fmt.Printf("All interview(s) are completed. Recording written to \"%s\".\n", recordConfig.replayFile.Name())
}

func runProxy(
	firstURL string,
	handleRequest func(*http.Request),
	redirectIfNeeded func(http.ResponseWriter, *http.Request),
	shouldCloseServer func(url string) bool) {
	var pendingRequestWaitGroup sync.WaitGroup
	var serverWaitGroup sync.WaitGroup

	client := http.Client{
		Timeout: globalConfig.requestTimeout,
	}

	var remoteHost string

	http.HandleFunc("/", func(response http.ResponseWriter, request *http.Request) {
		pendingRequestWaitGroup.Add(1)
		defer pendingRequestWaitGroup.Done()

		if request.URL.Path == "/" && request.Method == "GET" {
			// first request, so load the first URL and redirect
			httpResp, err := client.Get(firstURL)

			if err != nil {
				printError(err)
				return
			}

			http.Redirect(response, request, httpResp.Request.URL.Path, 302)

			remoteURL, _ := url.Parse(httpResp.Request.URL.String())
			remoteURL.Path = "/"
			remoteHost = remoteURL.String()
			return
		}

		handleRequest(request)

		remoteRequestURL, _ := url.Parse(remoteHost)
		remoteRequestURL.Path = request.URL.Path

		printVerbose("proxy", "INCOMING %s request on path %s\n", request.Method, request.URL.Path)
		printVerbose("proxy", "OUTGOING %s request to url %s\n", request.Method, remoteRequestURL)

		var httpResp *http.Response
		var err error

		switch request.Method {
		case "GET":
			httpResp, err = client.Get(remoteRequestURL.String())

			redirectIfNeeded(response, request)
		case "POST":
			request.ParseForm()
			form := request.Form
			httpResp, err = client.PostForm(remoteRequestURL.String(), form)

			http.Redirect(response, request, httpResp.Request.URL.Path, 302)
		}

		if err != nil {
			printError(err)
			return
		}

		headers := response.Header()

		for key, valList := range httpResp.Header {
			for _, val := range valList {
				headers.Add(key, val)
			}
		}

		// we don't want to cache anything!
		headers.Del("Cache-Control")

		response.Write(getBytesForHTTPResponse(*httpResp))

		if shouldCloseServer(request.URL.String()) {
			go func() {
				time.Sleep(500 * time.Millisecond)
				serverWaitGroup.Done()
			}()
		}
	})
	server := &http.Server{
		Addr: ":4222",
	}

	defer server.Close()

	serverWaitGroup.Add(1)

	url := "http://localhost" + server.Addr
	fmt.Printf("Serving on %s\n", url)

	go func() {
		server.ListenAndServe()
	}()

	openURLInBrowser(url)

	printVerbose("proxy", "Waiting for interview to finish...\n")
	serverWaitGroup.Wait()
	printVerbose("proxy", "Waiting for pending requests to finish...\n")
	pendingRequestWaitGroup.Wait()
	printVerbose("proxy", "Done, killing server now.\n")
}

func writeResponseToFile(outputFile *os.File, form url.Values) {
	for key, value := range form {
		if key != "screenId" {
			outputFile.WriteString(fmt.Sprintf("%s=%v\n", key, value))
		}
	}

	outputFile.WriteString("---\n")
}

func getBytesForHTTPResponse(response http.Response) []byte {
	defer response.Body.Close()

	buf := new(bytes.Buffer)
	buf.ReadFrom(response.Body)

	return buf.Bytes()
}

func openURLInBrowser(url string) {
	switch runtime.GOOS {
	case "linux":
		exec.Command("xdg-open", url).Start()
	case "windows":
		exec.Command("cmd", "/c", "start", url).Start()
	case "darwin":
		exec.Command("open", url).Start()
	}
}
