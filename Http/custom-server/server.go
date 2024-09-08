package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
)

type requestContextKey struct{}
type requestContextValue struct {
	requestID string
}

type appConfig struct {
	logger *log.Logger
}

type app struct {
	config  appConfig
	handler func(w http.ResponseWriter, r *http.Request, config appConfig) (int, error)
}

func (a app) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	status, err := a.handler(w, r, a.config)
	if err != nil {
		log.Printf("response_status=%d error=%s\n", status, err.Error())
		http.Error(w, err.Error(), status)
		return
	}
}

func idMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		ctx = context.WithValue(ctx, requestContextKey{}, requestContextValue{requestID: uuid.NewString()})

		r = r.WithContext(ctx)
		h.ServeHTTP(w, r)
	})
}

func panicMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rValue := recover(); rValue != nil {
				log.Println("panic detected", rValue)
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprint(w, "Unexpected server error")
			}
		}()
		h.ServeHTTP(w, r)
	})
}

func loggingMiddleware(h http.Handler) http.Handler {
	var requestId string
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()
		h.ServeHTTP(w, r)

		ctx := r.Context()
		v := ctx.Value(requestContextKey{})
		if m, ok := v.(requestContextValue); ok {
			requestId = m.requestID
		}
		log.Printf("id=%s path=%s method=%s duration=%f", requestId, r.URL.Path, r.Method, time.Since(startTime).Seconds())
	})
}

func apiHandler(w http.ResponseWriter, r *http.Request, config appConfig) (int, error) {
	config.logger.Println("Handling API request")
	fmt.Fprintf(w, "Hello, world!")
	return http.StatusOK, nil
}

func healthCheckHandler(w http.ResponseWriter, r *http.Request, config appConfig) (int, error) {
	if r.Method != http.MethodGet {
		return http.StatusMethodNotAllowed, fmt.Errorf("invalid request method:%s", r.Method)
	}
	config.logger.Println("Handling healthcheck request")
	fmt.Fprintf(w, "ok")
	return http.StatusOK, nil
}

func panicHandler(w http.ResponseWriter, r *http.Request, config appConfig) (int, error) {
	panic("I panicked")
}

func setupHandlers(mux *http.ServeMux, config appConfig) {
	mux.Handle("/health", app{config: config, handler: healthCheckHandler})
	mux.Handle("/api", app{config: config, handler: apiHandler})
	mux.Handle("/panic", app{config: config, handler: panicHandler})
}

func main() {
	listenAddr := os.Getenv("LISTEN_ADDR")
	if len(listenAddr) == 0 {
		listenAddr = ":8080"
	}

	config := appConfig{
		logger: log.New(os.Stdout, "", log.LstdFlags),
	}

	mux := http.NewServeMux()
	setupHandlers(mux, config)
	m := idMiddleware(loggingMiddleware(panicMiddleware(mux)))

	log.Fatal(http.ListenAndServe(listenAddr, m))
}
