package main

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"strings"
	"testing"
	"time"
)

func TestInputNoTimeout(t *testing.T) {
	r := strings.NewReader("jane")
	w := new(bytes.Buffer)
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	name, err := getNameContext(ctx, r, w)
	if err != nil {
		t.Errorf("Expected nil error, got %v", err)
	}
	if name != "jane" {
		t.Errorf("Expected name returned to be jane, got %s", name)
	}
}

func TestInputTimeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	r, _ := io.Pipe()
	name, err := getNameContext(ctx, r, os.Stdout)
	if err == nil {
		t.Error("Expected non-nil error, got nil")
	}

	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("Expected error: context.DeadlineExceeded, Got: %s", err)
	}

	if name != "default name" {
		t.Fatalf("Expected name returned to be Default Name, got %s", name)
	}
}
