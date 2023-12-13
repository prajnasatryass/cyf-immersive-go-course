package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"
)

func doHttpRequest(req *http.Request) *http.Response {
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failure doing HTTP request: %s\n", err)
		os.Exit(1)
	}
	return resp
}

func main() {
	req, err := http.NewRequest(http.MethodGet, "http://localhost:8080", nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failure creating HTTP request: %s\n", err)
		os.Exit(1)
	}

	var resp *http.Response
	resp = doHttpRequest(req)

	switch resp.StatusCode {
	case http.StatusTooManyRequests:
		// Assume Retry-After header is int
		var retryAfter int
		retryAfterStr := resp.Header.Get("Retry-After")
		retryAfter, err = strconv.Atoi(retryAfterStr)
		if err != nil {
			// Assume Retry-After header is in timestamp
			t, err := time.Parse(time.RFC1123, retryAfterStr)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Retry-After header is invalid: %s\n", err)
				os.Exit(1)
			}
			retryAfter = int(t.Sub(time.Now()).Seconds())
		}
		if retryAfter > 5 {
			fmt.Fprintf(os.Stdout, "Weather service is very busy. Please try again later")
			os.Exit(1)
		} else {
			fmt.Fprintf(os.Stdout, "Weather service is busy. Retrying in %d second(s)\n", retryAfter)
			time.Sleep(time.Duration(retryAfter) * time.Second)
		}

		resp = doHttpRequest(req)
	}

	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failure reading HTTP response: %s\n", err)
		os.Exit(1)
	}

	fmt.Println(string(respBody))
	return
}
