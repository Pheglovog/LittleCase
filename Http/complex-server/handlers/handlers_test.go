package handlers

import (
	"bytes"
	"complex-server/config"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func Test_apiHandler(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/api", nil)
	w := httptest.NewRecorder()

	b := new(bytes.Buffer)
	c := config.InitConfig(b)

	apiHandler(w, r, c)

	resp := w.Result()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected response status: %v, Got: %v\n", http.StatusOK, resp.StatusCode)
	}

	if string(body) != "Hello, world!" {
		t.Errorf("Expected response: Hello, world!, Got: %s\n", string(body))
	}
}

func Test_healthCheckHandler(t *testing.T) {
	tests := []struct {
		name         string
		method       string
		status       int
		responseBody string
	}{
		{
			name:         "test1",
			method:       http.MethodGet,
			status:       http.StatusOK,
			responseBody: "ok",
		},
		{
			name:         "test2",
			method:       http.MethodPost,
			status:       http.StatusMethodNotAllowed,
			responseBody: "Method not allowed\n",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := httptest.NewRequest(tc.method, "/health", nil)
			w := httptest.NewRecorder()

			b := new(bytes.Buffer)
			c := config.InitConfig(b)

			healthCheckHandler(w, r, c)

			resp := w.Result()
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Error(err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != tc.status {
				t.Errorf("Expected response status: %v, Got: %v\n", tc.status, resp.StatusCode)
			}

			if string(body) != tc.responseBody {
				t.Errorf("Expected response: %s, Got: %s\n", tc.responseBody, string(body))
			}

		})
	}

}
