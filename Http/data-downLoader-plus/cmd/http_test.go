package cmd

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func startTestHttpServer() *httptest.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/download", func(w http.ResponseWriter, req *http.Request) {
		fmt.Fprintf(w, "this is a response")
	})
	mux.HandleFunc("/upload", func(w http.ResponseWriter, req *http.Request) {
		defer req.Body.Close()
		data, err := io.ReadAll(req.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Fprintf(w, "JSON request received: %d bytes", len(data))
	})
	return httptest.NewServer(mux)
}

func TestHandleHttp(t *testing.T) {
	usageMessage := `
http: A HTTP client.
 
http: <options> server

Options: 
  -body string
    	JSON data for HTTP POST request
  -body-file string
    	File containing JSON data for HTTP POST request
  -output string
    	File path to write the response into
  -verb string
    	HTTP method (default "GET")
`
	ts := startTestHttpServer()
	defer ts.Close()

	outputFile := filepath.Join(t.TempDir(), "file_path.out")
	jsonBody := `{"id":1}`
	jsonBodyFile := filepath.Join(t.TempDir(), "data.json")

	err := os.WriteFile(jsonBodyFile, []byte(jsonBody), 0666)
	if err != nil {
		t.Fatal(err)
	}

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
		{
			name:   "test5",
			args:   []string{"-verb", "GET", "-output", outputFile, ts.URL + "/download"},
			errMsg: "",
			output: fmt.Sprintf("Data saved to: %s\n", outputFile),
		},
		{
			name:   "test6",
			args:   []string{"-verb", "POST", "-body", "", ts.URL + "/upload"},
			errMsg: ErrInvalidHTTPPostRequest.Error(),
			output: "Http POST request must specify a non-empty JSON body",
		},
		{
			name:   "test7",
			args:   []string{"-verb", "POST", "-body", jsonBody, ts.URL + "/upload"},
			errMsg: "",
			output: fmt.Sprintf("JSON request received: %d bytes\n", len(jsonBody)),
		},
		{
			name:   "test8",
			args:   []string{"-verb", "POST", "-body-file", jsonBodyFile, ts.URL + "/upload"},
			errMsg: "",
			output: fmt.Sprintf("JSON request received: %d bytes\n", len(jsonBody)),
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
