package cmd

import (
	"bytes"
	"flag"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func startTestHttpServer() *httptest.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/download", func(w http.ResponseWriter, req *http.Request) {
		fmt.Fprintf(w, "this is a response")
	})
	return httptest.NewServer(mux)
}

func TestHandleHttp(t *testing.T) {
	usageMessage := `
http: A HTTP client.
 
http: <options> server

Options: 
  -verb string
    	HTTP method (default "GET")
`
	ts := startTestHttpServer()
	defer ts.Close()

	tests := []struct {
		name   string
		args   []string
		output string
		errMsg string
	}{
		{
			name:   "test1",
			args:   []string{},
			output: "",
			errMsg: ErrNoServerSpecified.Error(),
		},
		{
			name:   "test2",
			args:   []string{"-h"},
			output: usageMessage,
			errMsg: flag.ErrHelp.Error(),
		},
		{
			name:   "test3",
			args:   []string{ts.URL + "/download"},
			errMsg: "",
			output: "this is a response\n",
		},
		{
			name:   "test4",
			args:   []string{"-verb", "PUT", "http://localhost"},
			errMsg: ErrInvalidHTTPMethod.Error(),
			output: "invalid HTTP method",
		},
	}

	w := new(bytes.Buffer)
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := HandleHttp(w, tc.args)
			var errMsg string
			if err != nil {
				errMsg = err.Error()
			}
			if errMsg != tc.errMsg {
				t.Errorf("Expected error message `%s`, got `%s`", tc.errMsg, errMsg)
			}

			output := w.String()
			if diff := cmp.Diff(output, tc.output); diff != "" {
				t.Errorf(diff)
			}
		})
		w.Reset()
	}
}
