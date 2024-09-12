package cmd

import (
	"bytes"
	"flag"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestHandleGrpc(t *testing.T) {
	usageMessage := `
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
			args:   []string{"-method", "service.host.local/method", "-body", "{}", "http://localhost"},
			output: "Executing grpc command\n",
			errMsg: "",
		},
	}

	w := new(bytes.Buffer)
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := HandleGrpc(w, tc.args)
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
