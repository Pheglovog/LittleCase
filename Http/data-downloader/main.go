package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

func fetchRemoteResource(client *http.Client, url string) ([]byte, error) {
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

func redirectPolicyFunc(req *http.Request, via []*http.Request) error {
	if len(via) >= 1 {
		return errors.New("stopped after 1 redirect")
	}
	return nil
}

func createHTTPCLientWithTimeout(d time.Duration) *http.Client {
	client := http.Client{Timeout: d, CheckRedirect: redirectPolicyFunc}
	return &client
}

func main() {
	if len(os.Args) != 2 {
		log.Fatal("Must specify a HTTP URL to get data from")
	}

	client := createHTTPCLientWithTimeout(15 * time.Second)
	body, err := fetchRemoteResource(client, os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	fmt.Fprintf(os.Stdout, "%s\n", body)
}
