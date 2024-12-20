package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"subCommand_improve/cmd"
)

var errInvalidSubCommand = errors.New("invalid sub-command specified")

func printUsage(w io.Writer) {
	fmt.Fprint(w, "Usage: mync [http|grpc] -h\n")
	cmd.HandleHttp(w, []string{"-h"})
	cmd.HandleGrpc(w, []string{"-h"})
}

func handleCommand(w io.Writer, args []string) error {
	var err error

	if len(args) < 1 {
		err = errInvalidSubCommand
	} else {
		switch args[0] {
		case "http":
			err = cmd.HandleHttp(w, args[1:])
		case "grpc":
			err = cmd.HandleGrpc(w, args[1:])
		case "-h", "-help":
			printUsage(w)
		default:
			err = errInvalidSubCommand
		}
	}
	if errors.Is(err, cmd.ErrNoServerSpecified) || errors.Is(err, errInvalidSubCommand) {
		fmt.Fprintln(w, err)
		printUsage(w)
	}
	return err
}

func main() {
	err := handleCommand(os.Stdout, os.Args[1:])
	if err != nil {
		os.Exit(1)
	}
}
