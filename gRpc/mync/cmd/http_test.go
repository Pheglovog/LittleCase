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
	"strings"
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
	mux.HandleFunc("/redirect", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "new-url", http.StatusMovedPermanently)
	})
	mux.HandleFunc("/debug-header-response", func(w http.ResponseWriter, r *http.Request) {
		headers := []string{}
		for k, v := range r.Header {
			if strings.HasPrefix(k, "Debug") {
				headers = append(headers, fmt.Sprintf("%s=%s", k, v[0]))
			}
		}
		fmt.Fprint(w, strings.Join(headers, " "))
	})
	mux.HandleFunc("/debug-basicauth", func(w http.ResponseWriter, r *http.Request) {
		u, p, ok := r.BasicAuth()
		if !ok {
			http.Error(w, "Basic auth missing/malformed", http.StatusBadRequest)
			return
		}
		fmt.Fprintf(w, "%s=%s", u, p)
	})
	return httptest.NewServer(mux)
}

func TestHandleHttp(t *testing.T) {
	usageMessage := `
http: A HTTP client.
 
http: <options> server

Options: 
  -basicAuth string
    	Add basic auth (username:password) credentials to the outgoing request
  -body string
    	JSON data for HTTP POST request
  -body-file string
    	File containing JSON data for HTTP POST request
  -disable-redirect
    	Do not follow redirection request
  -header value
    	Add one or more headers to the outgoing request (key=value)
  -max-idle-conns int
    	Maximum number of idle connections for the connection pool
  -num-requests int
    	Number of requests to make (default 1)
  -output string
    	File path to write the response into
  -report
    	report this http request's latency
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
			output: "",
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
			output: "",
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
		{
			name:   "test9",
			args:   []string{"-disable-redirect", ts.URL + "/redirect"},
			errMsg: `Get "/new-url": stopped after 1 redirect`,
			output: "",
		},
		{
			name:   "test10",
			args:   []string{"-header", "Debug-Key1=value1", "-header", "Debug-Key2=value2", ts.URL + "/debug-header-response"},
			errMsg: "",
			output: "Debug-Key1=value1 Debug-Key2=value2\n",
		},
		{
			name:   "test11",
			args:   []string{"-basicAuth", "user=password", ts.URL + "/debug-basicauth"},
			errMsg: "",
			output: "user=password\n",
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
