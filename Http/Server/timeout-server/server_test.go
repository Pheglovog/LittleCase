package main

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func Test_handleUserAPI(t *testing.T) {
	var err error
	logBuf := new(bytes.Buffer)
	timeoutDuration := 30 * time.Second
	logger := log.New(logBuf, "", log.LstdFlags)

	mux := http.NewServeMux()
	setupHandlers(mux, logger)
	mTimeout := http.TimeoutHandler(mux, timeoutDuration, "I ran out of time")

	ts := httptest.NewServer(mTimeout)
	defer ts.Close()

	client := http.Client{
		Timeout: 4 * time.Second,
	}
	_, err = client.Get(ts.URL + "/api/users/" + "?ping_server=" + ts.URL)
	if err == nil {
		t.Fatal("Expected no-nil error")
	}
	time.Sleep(1 * time.Second)
	expectedServerLogLine := "Aborting request processing: context canceled"
	if !strings.Contains(logBuf.String(), expectedServerLogLine) {
		t.Fatalf("Expected server log to contain: %s\n Got: %s", expectedServerLogLine, logBuf.String())
	}
}
