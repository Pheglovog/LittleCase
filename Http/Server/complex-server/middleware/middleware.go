package middleware

import (
	"complex-server/config"
	"fmt"
	"net/http"
	"time"
)

func panicMiddleware(h http.Handler, conf config.AppConfig) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rValue := recover(); rValue != nil {
					conf.Logger.Println("panic detected", rValue)
					w.WriteHeader(http.StatusInternalServerError)
					fmt.Fprint(w, "Unexpected server error")
				}
			}()
			h.ServeHTTP(w, r)
		})
}

func loggingMiddleware(h http.Handler, conf config.AppConfig) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()
		h.ServeHTTP(w, r)

		conf.Logger.Printf("protocol=%s path=%s method=%s duration=%f", r.Proto, r.URL.Path, r.Method, time.Since(startTime).Seconds())
	})
}
