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
	var interviewWaitGroup sync.WaitGroup
	var pendingRequestsWaitGroup sync.WaitGroup

	client := http.Client{
		Timeout: globalConfig.requestTimeout,
	}
	lastURL := recordConfig.interviewURL

	handleProxyRequest := func(response http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/" {
			pendingRequestsWaitGroup.Add(1)
			newURL, _ := url.Parse(lastURL)
			newURL.Path = request.URL.Path

			printVerbose("server", "static request: %v\n", newURL.Path)
			httpResp, err := client.Get(newURL.String())

			if err != nil {
				fmt.Printf("ERROR %v\n", err)
				return
			}

			headers := response.Header()

			for key, valList := range httpResp.Header {
				for _, val := range valList {
					headers.Add(key, val)
				}
			}

			response.Write(getBytesForHTTPResponse(*httpResp))
			pendingRequestsWaitGroup.Done()
			return
		}

		printVerbose("server", "page request: %s\n", request.Method)

		switch request.Method {
		case "GET":
			httpResp, _ := client.Get(lastURL)

			lastURL = httpResp.Request.URL.String()
			response.Write(getBytesForHTTPResponse(*httpResp))
		case "POST":
			request.ParseForm()
			form := request.Form

			printVerbose("server", "Body: %v\n", form)
			writeResponseToFile(recordConfig.outputFile, form)
			httpResp, _ := client.PostForm(lastURL, form)

			lastURL = httpResp.Request.URL.String()
			response.Write(getBytesForHTTPResponse(*httpResp))

			if strings.Contains(lastURL, endOfInterviewPath) {
				go func() {
					time.Sleep(250 * time.Millisecond)
					interviewWaitGroup.Done()
				}()
			}
		}
	}

	http.HandleFunc("/", handleProxyRequest)
	server := &http.Server{
		Addr: ":8080",
	}

	defer server.Close()

	interviewWaitGroup.Add(1)
	go func() {
		server.ListenAndServe()
	}()

	url := "http://localhost" + server.Addr
	fmt.Printf("Serving on %s\n", url)
	openURLInBrowser(url)

	interviewWaitGroup.Wait()
	pendingRequestsWaitGroup.Wait()

	fmt.Printf("Completed interview. Recording written to \"%s\".\n", recordConfig.outputFile.Name())
}

func handleStaticRequest(response http.ResponseWriter, request *http.Request, baseURL string) {

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
