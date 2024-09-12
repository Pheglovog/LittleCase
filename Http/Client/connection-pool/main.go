package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptrace"
	"os"
	"time"
)

func createHTTPGetRequestWithTrace(ctx context.Context, url string) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	trace := &httptrace.ClientTrace{
		DNSDone: func(di httptrace.DNSDoneInfo) {
			fmt.Printf("DNS Info: %+v\n", di)
		},
		GotConn: func(gci httptrace.GotConnInfo) {
			fmt.Printf("Got Conn: %+v\n", gci)
		},
		TLSHandshakeStart: func() {
			fmt.Printf("TLS HandShake Start\n")
		},
		TLSHandshakeDone: func(connState tls.ConnectionState, err error) {
			fmt.Printf("TLS HandShake Done\n")
		},

		PutIdleConn: func(err error) {
			fmt.Printf("Put Idle Conn Error: %+v\n", err)
		},
	}

	ctxTrace := httptrace.WithClientTrace(req.Context(), trace)
	req = req.WithContext(ctxTrace)
	return req, nil
}

func createHTTPClientWithTimeout(d time.Duration) *http.Client {
	client := http.Client{Timeout: d}
	return &client
}

func main() {
	d := 5 * time.Second
	ctx := context.Background()
	client := createHTTPClientWithTimeout(d)

	req, err := createHTTPGetRequestWithTrace(ctx, os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	for {
		resp, _ := client.Do(req)
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
		fmt.Printf("Resp protocol: %#v\n", resp.Proto)
		time.Sleep(1 * time.Second)
		fmt.Println("------------")
	}
}
