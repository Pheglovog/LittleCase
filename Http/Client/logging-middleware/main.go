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

type LoggingClient struct {
	log *log.Logger
}

func (c LoggingClient) RoundTrip(r *http.Request) (*http.Response, error) {
	c.log.Printf("Sending a %s request to %s over %s\n", r.Method, r.URL, r.Proto)
	resp, err := http.DefaultTransport.RoundTrip(r)
	c.log.Printf("Got back a response over %s\n", resp.Proto)

	return resp, err
}

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

	l := log.New(os.Stdout, "", log.LstdFlags)
	myTransport := LoggingClient{log: l}

	client := createHTTPCLientWithTimeout(15 * time.Second)
	client.Transport = &myTransport

	body, err := fetchRemoteResource(client, os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	fmt.Fprintf(os.Stdout, "Bytes in response: %d\n", len(body))
}
