package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptrace"
	"os"
	"time"
)

func handlePing(w http.ResponseWriter, r *http.Request) {
	log.Println("ping: Got a request")
	time.Sleep(3 * time.Second)
	fmt.Fprint(w, "pong")

}

func createHTTPGetRequestWithTrace(ctx context.Context, url string) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	trace := &httptrace.ClientTrace{
		DNSStart: func(di httptrace.DNSStartInfo) {
			fmt.Printf("httptrace.DNSDoneInfo: %v\n", di)
		},
		DNSDone: func(di httptrace.DNSDoneInfo) {
			fmt.Printf("httptrace.DNSDoneInfo: %v\n", di)
		},
		GotConn: func(gci httptrace.GotConnInfo) {
			fmt.Printf("httptrace.GotConnInfo: %v\n", gci)
		},
		TLSHandshakeStart: func() {
			fmt.Print("TLS HandShake Done\n")
		},
		PutIdleConn: func(err error) {
			fmt.Printf("Put Idle conn Error: %v\n", err)
		},
	}
	ctxTrace := httptrace.WithClientTrace(req.Context(), trace)
	req = req.WithContext(ctxTrace)
	return req, err
}

func handleUserAPI(logger *log.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		pingServer := r.URL.Query().Get("ping_server")
		if len(pingServer) == 0 {
			pingServer = "http://localhost:8080"
		}

		done := make(chan struct{})
		logger.Println("I started processing the request")

		req, err := createHTTPGetRequestWithTrace(r.Context(), pingServer+"/ping")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		client := &http.Client{}
		logger.Println("Outgoing HTTP request")
		resp, err := client.Do(req)
		if err != nil {
			logger.Printf("Error making request: %v\n", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()
		data, _ := io.ReadAll(resp.Body)

		logger.Println("Processing the response i got")

		go func() {
			time.Sleep(5 * time.Second)
			close(done)
		}()

		select {
		case <-done:
			logger.Println("doSomeWork done: Continuing request processing")
		case <-r.Context().Done():
			logger.Printf("Aborting request processing: %v\n", r.Context().Err())
			return
		}

		fmt.Fprint(w, string(data))
		logger.Println("I finished processing the request")
	}
}

func setupHandlers(mux *http.ServeMux, logger *log.Logger) {
	userHandler := handleUserAPI(logger)
	mux.Handle("/api/users/", userHandler)
	mux.HandleFunc("/ping", handlePing)
}

func main() {
	listenAddr := os.Getenv("LISTEN_ADDR")
	if len(listenAddr) == 0 {
		listenAddr = ":8080"
	}

	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime|log.Lshortfile)
	timeoutDuration := 30 * time.Second
	mux := http.NewServeMux()
	setupHandlers(mux, logger)
	mTimeout := http.TimeoutHandler(mux, timeoutDuration, "I ran out of time")

	s := http.Server{
		Addr:         listenAddr,
		Handler:      mTimeout,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}
	log.Fatal(s.ListenAndServe())
}
