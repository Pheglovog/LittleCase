package main

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func Test_setupHandlers(t *testing.T) {
	buf := new(bytes.Buffer)
	mux := http.NewServeMux()
	l := log.New(buf, "test-tls-server", log.LstdFlags)

	m := setupHandlers(mux, l)

	ts := httptest.NewUnstartedServer(m)
	ts.EnableHTTP2 = true
	ts.StartTLS()

	client := ts.Client()
	_, err := client.Get(ts.URL + "/api")
	if err != nil {
		t.Fatal(err)
	}
	expected := "protocol=HTTP/2.0 path=/api method=GET"
	mLogs := buf.String()

	if !strings.Contains(mLogs, expected) {
		t.Errorf("Expected logs to contain %s, Found: %s\n", expected, mLogs)
	}
}
