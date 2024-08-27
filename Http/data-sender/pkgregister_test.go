package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func packageRegHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		p := pkgData{}
		d := pkgRegisterResult{}
		defer r.Body.Close()
		data, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}
		err = json.Unmarshal(data, &p)
		if err != nil || p.Name == "" || p.Version == "" {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		d.Id = p.Name + "-" + p.Version
		jsonData, err := json.Marshal(d)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, string(jsonData))
	} else {
		http.Error(w, "Invalid HTTP method specified", http.StatusMethodNotAllowed)
		return
	}
}

func startTestPackageServer() *httptest.Server {
	ts := httptest.NewServer(http.HandlerFunc(packageRegHandler))
	return ts
}

func Test_registerPackageData(t *testing.T) {
	ts := startTestPackageServer()
	defer ts.Close()

	p := pkgData{
		Name:    "MyPackage",
		Version: "0.1",
	}

	resp, err := registerPackageData(ts.URL, p)
	if err != nil {
		t.Error(err)
	}
	if resp.Id != "MyPackage-0.1" {
		t.Errorf("Expected package id to be MyPackage-0.1, Got: %s", resp.Id)
	}
}

func Test_registerEmptyPackageData(t *testing.T) {
	ts := startTestPackageServer()
	ts.Close()

	resp, err := registerPackageData(ts.URL, pkgData{})
	if err == nil {
		t.Error("Expected error to be non-nil, got nil")
	}
	if resp.Id != "" {
		t.Errorf("Expected package ID to be empty, got: %s", resp.Id)
	}
}
