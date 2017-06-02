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

func startProxyForInterview(outputFile *os.File) {
	var completedInterview sync.WaitGroup

	client := http.Client{
		Timeout: globalConfig.requestTimeout,
	}
	lastURL := recordConfig.interviewURL

	handleProxyRequest := func(response http.ResponseWriter, request *http.Request) {
		printVerbose("server", "New %s request\n", request.Method)
		switch request.Method {
		case "GET":
			request.ParseForm()

			httpResp, _ := client.Get(lastURL)

			lastURL = httpResp.Request.URL.String()
			response.Write(getBytesForHTTPResponse(*httpResp))
		case "POST":
			request.ParseForm()
			form := request.Form

			printVerbose("server", "Body: %v\n", form)
			writeResponseToFile(outputFile, form)
			httpResp, _ := client.PostForm(lastURL, form)

			lastURL = httpResp.Request.URL.String()
			response.Write(getBytesForHTTPResponse(*httpResp))

			if strings.Contains(lastURL, endOfInterviewPath) {
				go func() {
					time.Sleep(250 * time.Millisecond)
					completedInterview.Done()
				}()
			}
		}
	}

	http.HandleFunc("/", handleProxyRequest)
	server := &http.Server{
		Addr: ":8080",
	}

	defer server.Close()

	completedInterview.Add(1)
	go func() {
		server.ListenAndServe()
	}()

	url := "http://localhost" + server.Addr
	fmt.Printf("Serving on %s\n", url)
	openURLInBrowser(url)

	completedInterview.Wait()
	fmt.Printf("Completed interview. Recording written to \"%s\".\n", recordConfig.outputFile)
}

func writeResponseToFile(outputFile *os.File, form url.Values) {
	outputFile.WriteString("---\n")

	for key, value := range form {
		if key != "screenId" {
			outputFile.WriteString(fmt.Sprintf("%s=%v\n", key, value))
		}
	}
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
