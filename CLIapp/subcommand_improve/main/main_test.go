package main

import (
	"bytes"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func Test_handleCommnd(t *testing.T) {
	usageMessage := `Usage: mync [http|grpc] -h

http: A HTTP client.
 
http: <options> server

Options: 
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
