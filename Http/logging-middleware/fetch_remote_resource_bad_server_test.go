package main

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func startBadTestHTTPServerV2(shutdownServer chan struct{}) *httptest.Server {
	ts := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				<-shutdownServer
				fmt.Fprint(w, "Hello, World")
			},
		),
	)
	return ts
}

func Test_fetchBadRemoteResourceV2(t *testing.T) {
	shutdownServer := make(chan struct{})

	ts := startBadTestHTTPServerV2(shutdownServer)
	defer ts.Close()
	defer func() {
		shutdownServer <- struct{}{}
	}()

	client := createHTTPCLientWithTimeout(200 * time.Millisecond)
	_, err := fetchRemoteResource(client, ts.URL)
	if err == nil {
		t.Fatal("Expected non-nil error")
	}
	if !strings.Contains(err.Error(), context.DeadlineExceeded.Error()) {
		t.Fatalf("Expected error to contain : %v, Got: %v", context.DeadlineExceeded, err)
	}
}
