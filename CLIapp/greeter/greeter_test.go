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
		errMsg string
		config config
	}{
		{
			name:   "test1",
			args:   []string{"-h"},
			errMsg: "",
			config: config{PrintUsage: true, NumTimes: 0},
		},
		{
			name:   "test2",
			args:   []string{"10"},
			errMsg: "",
			config: config{PrintUsage: false, NumTimes: 10},
		},
		{
			name:   "test3",
			args:   []string{"abc"},
			errMsg: "strconv.Atoi: parsing \"abc\": invalid syntax",
			config: config{PrintUsage: false, NumTimes: 0},
		},
		{
			name:   "test4",
			args:   []string{"1", "foo"},
			errMsg: "invalid number of arguments",
			config: config{PrintUsage: false, NumTimes: 0},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			config, err := parseArgs(tc.args)
			if diff := cmp.Diff(tc.config, config); diff != "" {
				t.Error(diff)
			}

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
			name:   "test1",
			config: config{PrintUsage: true, NumTimes: 0},
			in:     "",
			out:    usageString,
			errMsg: "",
		},
		{
			name:   "test2",
			config: config{PrintUsage: false, NumTimes: 5},
			in:     "",
			out:    "Your name Please?\n",
			errMsg: "you didn't enter your name",
		},
		{
			name:   "test3",
			config: config{PrintUsage: false, NumTimes: 5},
			in:     "Bill",
			out:    "Your name Please?\n" + strings.Repeat("Nice to meet you Bill\n", 5),
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
