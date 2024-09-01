package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func startTestHttpServer() *httptest.Server {
	ts := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprint(w, "Hello World")
			},
		),
	)
	return ts
}

func Test_fetchRemoteResource(t *testing.T) {
	ts := startTestHttpServer()
	defer ts.Close()

	expected := "Hello World"

	data, err := fetchRemoteResource(http.DefaultClient, ts.URL)
	if err != nil {
		t.Error(err)
	}
	if expected != string(data) {
		t.Errorf("Expected response to be: %s, Got: %s", expected, data)
	}
}
