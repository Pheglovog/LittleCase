package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"
)

func createContextWithTimeout(d time.Duration) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithTimeout(context.Background(), d)
	return ctx, cancel
}

func setupSignalHandler(w io.Writer, cancelFunc context.CancelFunc) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		s := <-c
		fmt.Fprintf(w, "Got signal:%v\n", s)
		cancelFunc()
	}()
}

func executeCommand(ctx context.Context, command string, arg string) error {
	return exec.CommandContext(ctx, command, arg).Run()
}

func main() {
	if len(os.Args) != 3 {
		log.Fatalf("Usage: %s <command> <argument>\n", os.Args[0])
	}

	command := os.Args[1]
	arg := os.Args[2]
	cmdTimeout := 10 * time.Second
	ctx, cancel := createContextWithTimeout(cmdTimeout)
	defer cancel()

	setupSignalHandler(os.Stdout, cancel)

	err := executeCommand(ctx, command, arg)
	if err != nil {
		log.Fatal(err)
	}
}
