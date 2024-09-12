package middleware

import (
	"bytes"
	"complex-server/config"
	"complex-server/handlers"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func Test_panicMiddleware(t *testing.T) {
	b := new(bytes.Buffer)
	c := config.InitConfig(b)

	mux := http.NewServeMux()
	handlers.Register(mux, c)

	h := panicMiddleware(mux, c)

	r := httptest.NewRequest(http.MethodGet, "/panic", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)

	resp := w.Result()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("Expected response status: %v, Got: %v\n", http.StatusOK, resp.StatusCode)
	}

	if string(body) != "Unexpected server error" {
		t.Errorf("Expected response: Unexpected server error, Got: %s\n", string(body))
	}
}
