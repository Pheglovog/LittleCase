package main

import (
	"bytes"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func Test_parseArgs(t *testing.T) {
	tests := []struct {
		name   string
		args   []string
		config config
		output string
		errMsg string
	}{
		{
			name:   "test1",
			args:   []string{"-h"},
			config: config{NumTimes: 0},
			output: `
A greeter application which prints the name you entered a specified number of times.
 
Usage of greeter: <option> [name]
  -n int
        Number of times to greet`,
			errMsg: "flag: help requested",
		},
		{
			name:   "test2",
			args:   []string{"-n", "10"},
			config: config{NumTimes: 10},
			errMsg: "",
		},
		{
			name:   "test3",
			args:   []string{"-n", "abc"},
			config: config{NumTimes: 0},
			errMsg: "invalid value \"abc\" for flag -n: parse error",
		},
		{
			name:   "test4",
			args:   []string{"-n", "1", "John"},
			config: config{NumTimes: 1, Name: "John"},
			errMsg: "",
		},
		{
			name:   "test5",
			args:   []string{"-n", "1", "John", "Doe"},
			config: config{NumTimes: 1},
			errMsg: "more than one positional argument specified",
		},
	}

	w := new(bytes.Buffer)
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			c, err := parseArgs(w, tc.args)
			if diff := cmp.Diff(c, tc.config); diff != "" {
				t.Errorf(diff)
			}

			var errMsg string
			if err != nil {
				errMsg = err.Error()
			}
			if errMsg != tc.errMsg {
				t.Errorf("Expected error message `%s`, got `%s`", tc.errMsg, errMsg)
			}
		})
		w.Reset()
	}
}

func Test_validateArgs(t *testing.T) {
	tests := []struct {
		name   string
		config config
		errMsg string
	}{
		{
			name:   "test1",
			config: config{},
			errMsg: "must specify a number greater than 0",
		},
		{
			name:   "test2",
			config: config{NumTimes: -1},
			errMsg: "must specify a number greater than 0",
		},
		{
			name:   "test3",
			config: config{NumTimes: 10},
			errMsg: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := validateArgs(tc.config)

			var errMsg string
			if err != nil {
				errMsg = err.Error()
			}
			if errMsg != tc.errMsg {
				t.Errorf("Expected error message `%s`, got `%s`", tc.errMsg, errMsg)
			}
		})
	}
}

func Test_runCmd(t *testing.T) {
	tests := []struct {
		name   string
		config config
		in     string
		out    string
		errMsg string
	}{
		{
			name:   "test2",
			config: config{NumTimes: 5},
			in:     "",
			out:    "Your name Please?\n",
			errMsg: "you didn't enter your name",
		},
		{
			name:   "test3",
			config: config{NumTimes: 5},
			in:     "Bill",
			out:    "Your name Please?\n" + strings.Repeat("Nice to meet you Bill\n", 5),
			errMsg: "",
		},
		{
			name:   "test4",
			config: config{NumTimes: 5, Name: "Bill"},
			in:     "",
			out:    strings.Repeat("Nice to meet you Bill\n", 5),
			errMsg: "",
		},
	}

	write := new(bytes.Buffer)
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			read := strings.NewReader(tc.in)
			err := runCmd(read, write, tc.config)
			out := write.String()
			if out != tc.out {
				t.Errorf("Expected output `%s`, got `%s`", tc.out, out)
			}
			var errMsg string
			if err != nil {
				errMsg = err.Error()
			}
			if errMsg != tc.errMsg {
				t.Errorf("Expected error message `%s`, got `%s`", tc.errMsg, errMsg)
			}
			write.Reset()
		})
	}
}
