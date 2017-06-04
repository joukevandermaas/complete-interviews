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
	client := http.Client{
		Timeout: globalConfig.requestTimeout,
	}
	resp, err := client.Get(recordConfig.interviewURL)

	if err != nil {
		printError(err)
		return
	}

	baseURL := resp.Request.URL.String()

	handleRequest := func(request *http.Request) {
		if request.Method == "POST" {
			request.ParseForm()
			form := request.Form

			printVerbose("recording", "Recording interview answer %v\n", form)
			writeResponseToFile(recordConfig.replayFile, form)
		}
	}

	isInterviewDone := func(url string) bool {
		return strings.Contains(url, endOfInterviewPath)
	}

	runProxy(baseURL, handleRequest, isInterviewDone)

	fmt.Printf("Completed interview. Recording written to \"%s\".\n", recordConfig.replayFile.Name())
}

func runProxy(baseURL string, handleRequest func(*http.Request), isDone func(string) bool) {
	var pendingRequestWaitGroup sync.WaitGroup
	var serverWaitGroup sync.WaitGroup

	client := http.Client{
		Timeout: globalConfig.requestTimeout,
	}

	http.HandleFunc("/", func(response http.ResponseWriter, request *http.Request) {
		pendingRequestWaitGroup.Add(1)
		remoteRequestURL := baseURL

		handleRequest(request)

		// For anything not on the root of our server (e.g. scripts/css/images),
		// we should use the path that was requested. For interview pages, we should
		// use the base url we've been given
		if request.URL.Path != "/" {
			newURL, _ := url.Parse(baseURL)
			newURL.Path = request.URL.Path

			remoteRequestURL = newURL.String()
		}

		printVerbose("proxy", "INCOMING %s request on path %s\n", request.Method, request.URL.Path)
		printVerbose("proxy", "OUTGOING %s request to url %s\n", request.Method, remoteRequestURL)

		var httpResp *http.Response
		var err error

		switch request.Method {
		case "GET":
			httpResp, err = client.Get(remoteRequestURL)
		case "POST":
			request.ParseForm()
			form := request.Form
			httpResp, err = client.PostForm(remoteRequestURL, form)
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
		pendingRequestWaitGroup.Done()

		if isDone(httpResp.Request.URL.String()) {
			go func() {
				printVerbose("proxy", "Done, killing server now.\n")
				time.Sleep(250 * time.Millisecond)
				serverWaitGroup.Done()
			}()
		}
	})
	server := &http.Server{
		Addr: ":8080",
	}

	defer server.Close()

	serverWaitGroup.Add(1)

	url := "http://localhost" + server.Addr
	fmt.Printf("Serving on %s\n", url)

	go func() {
		server.ListenAndServe()
	}()

	openURLInBrowser(url)

	serverWaitGroup.Wait()
	pendingRequestWaitGroup.Wait()
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
