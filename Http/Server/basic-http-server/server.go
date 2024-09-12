package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
)

type requestContextKey struct{}
type requestContextValue struct {
	requestID string
}

func addRequestID(r *http.Request, requestID string) *http.Request {
	c := requestContextValue{
		requestID: requestID,
	}
	currentCtx := r.Context()
	newCtx := context.WithValue(currentCtx, requestContextKey{}, c)
	return r.WithContext(newCtx)
}

type logLine struct {
	RequestID     string `json:"request-id"`
	URL           string `json:"url"`
	Method        string `json:"method"`
	ContentLength int64  `json:"content_length"`
	Protocol      string `json:"protocol"`
}

func logRequest(req *http.Request) {
	l := logLine{
		URL:           req.Host + req.URL.String(),
		Method:        req.Method,
		ContentLength: req.ContentLength,
		Protocol:      req.Proto,
	}

	ctx := req.Context()
	v := ctx.Value(requestContextKey{})
	if m, ok := v.(requestContextValue); ok {
		l.RequestID = m.requestID
	}

	data, err := json.Marshal(l)
	if err != nil {
		panic(err)
	}
	log.Println(string(data))
}

// func processRequest(w http.ResponseWriter, r *http.Request) {
// 	logRequest(r)
// 	fmt.Fprint(w, "Request processed")
// }

func apiHandler(w http.ResponseWriter, req *http.Request) {
	requestID := "request-123-abc"
	req = addRequestID(req, requestID)
	logRequest(req)
	fmt.Fprintf(w, "Hello, world!")
}

func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	logRequest(r)
	fmt.Fprintf(w, "ok")
}

func catchAllHandler(w http.ResponseWriter, r *http.Request) {
	logRequest(r)
	fmt.Fprint(w, "your request was processed by the catch all handler")
}

func setupHandlers(mux *http.ServeMux) {
	mux.HandleFunc("/health", healthCheckHandler)
	mux.HandleFunc("/api", apiHandler)
	mux.HandleFunc("/", catchAllHandler)
}

func main() {
	listenAddr := os.Getenv("LISTEN_ADDR")
	if len(listenAddr) == 0 {
		listenAddr = ":8080"
	}

	mux := http.NewServeMux()
	setupHandlers(mux)

	log.Fatal(http.ListenAndServe(listenAddr, mux))
}
