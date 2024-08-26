package cmd

import (
	"bytes"
	"flag"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestHandleHttp(t *testing.T) {
	usageMessage := `
http: A HTTP client.
 
http: <options> server

Options: 
  -verb string
    	HTTP method (default "GET")
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
			args:   []string{"http://localhost"},
			errMsg: "",
			output: "Executing http command\n",
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
				// t.Errorf("Expected `%s`, got `%s`", tc.output, output)
				t.Errorf(diff)
			}
		})
		w.Reset()
	}
}
