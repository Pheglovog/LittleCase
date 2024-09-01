package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func Test_handleCommnd(t *testing.T) {
	usageMessage := `Usage: mync [http|grpc] -h

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
  -output string
    	File path to write the response into
  -verb string
    	HTTP method (default "GET")

grpc: A gRPC client.
 
grpc: <options> server

Options:
  -body string
    	Body of request
  -method string
    	Method to call
`
	tests := []struct {
		name   string
		args   []string
		output string
		errMsg string
	}{
		{
			name:   "test1",
			args:   []string{},
			output: "invalid sub-command specified\n" + usageMessage,
			errMsg: "invalid sub-command specified",
		},
		{
			name:   "test2",
			args:   []string{"-h"},
			output: usageMessage,
			errMsg: "",
		},
		{
			name:   "test3",
			args:   []string{"foo"},
			output: "invalid sub-command specified\n" + usageMessage,
			errMsg: "invalid sub-command specified",
		},
	}

	w := new(bytes.Buffer)
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := handleCommand(w, tc.args)

			var errMsg string
			if err != nil {
				errMsg = err.Error()
			}
			if errMsg != tc.errMsg {
				t.Errorf("Expected error message `%s`, got `%s`", tc.errMsg, errMsg)
			}
			output := w.String()
			if diff := cmp.Diff(output, tc.output); diff != "" {
				// t.Errorf("Expected `%s`, got `%s`", tc.output, output)
				t.Errorf(diff)
			}
		})
		w.Reset()
	}
}

var binaryName string
var testServerURL string

func TestMain(m *testing.M) {
	if runtime.GOOS == "windows" {
		binaryName = "mync.exe"
	} else {
		binaryName = "mync"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2000*time.Millisecond)
	defer cancel()
	cmd := exec.CommandContext(ctx, "go", "build", "-o", binaryName)
	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}

	defer func() {
		err = os.Remove(binaryName)
		if err != nil {
			log.Fatalf("Error removing built binary: %v", err)
		}
	}()

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "this is a response")
	})
	ts := httptest.NewServer(mux)
	testServerURL = ts.URL
	defer ts.Close()
	m.Run()
}

func TestSubCommandInvoke(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	curDir, err := os.Getwd()
	if err != nil {
		t.Error(err)
	}
	binaryPath := path.Join(curDir, binaryName)
	t.Log(binaryPath)

	tests := []struct {
		name     string
		args     []string
		input    string
		output   []string
		exitCode int
	}{
		{
			name:     "test1",
			args:     []string{},
			input:    "",
			output:   []string{},
			exitCode: 1,
		},
		{
			name:     "test2",
			args:     []string{"http"},
			input:    "",
			output:   []string{"you have to specify the remote server"},
			exitCode: 1,
		},
		{
			name:     "test3",
			args:     []string{"http", testServerURL},
			input:    "",
			output:   []string{"this is a response"},
			exitCode: 0,
		},
		{
			name:     "test4",
			args:     []string{"http", "-verb", "POST", "-body", `"{"id":1}"`, testServerURL},
			input:    "",
			output:   []string{"this is a response"},
			exitCode: 0,
		},
		{
			name:     "test5",
			args:     []string{"http", "-method", "POST", testServerURL},
			input:    "",
			output:   []string{"flag provided but not defined: -method"},
			exitCode: 1,
		},
		{
			name:  "test6",
			args:  []string{"grpc"},
			input: "",
			output: []string{
				"you have to specify the remote server",
			},
			exitCode: 1,
		},
		{
			name:  "test7",
			args:  []string{"grpc", "127.0.0.1"},
			input: "",
			output: []string{
				"Executing grpc command",
			},
			exitCode: 0,
		},
	}

	w := new(bytes.Buffer)
	for _, tc := range tests {
		t.Logf("Executing:%v %v\n", binaryPath, tc.args)
		t.Run(tc.name, func(t *testing.T) {
			cmd := exec.CommandContext(ctx, binaryPath, tc.args...)
			cmd.Stdout = w

			if len(tc.input) != 0 {
				cmd.Stdin = strings.NewReader(tc.input)
			}
			err := cmd.Run()

			if err != nil && tc.exitCode == 0 {
				t.Errorf("Expected application to exit without an error. Got: %v", err)
			}

			if cmd.ProcessState.ExitCode() != tc.exitCode {
				t.Log(w.String())
				t.Errorf("Expected application to have exit code: %v. Got: %v", tc.exitCode, cmd.ProcessState.ExitCode())
			}

			output := w.String()
			lines := strings.Split(output, "\n")
			for num := range tc.output {
				if lines[num] != tc.output[num] {
					t.Errorf("Expected output line to be:%v, Got:%v", tc.output[num], lines[num])
				}
			}
		})
		w.Reset()
	}
}
