package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func startTestPackageServer() *httptest.Server {
	pkgData := `[
	{"name":"package1", "version":"1.1"},
	{"name":"package2", "version":"1.0"}
	]`

	ts := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, pkgData)
		}),
	)
	return ts
}

func Test_fetchPackageData(t *testing.T) {
	ts := startTestPackageServer()
	defer ts.Close()
	packages, err := fetchPackageData(ts.URL)
	if err != nil {
		t.Error(err)
	}
	if len(packages) != 2 {
		t.Errorf("Expected 2 packages, Got back: %d", len(packages))
	}
}
