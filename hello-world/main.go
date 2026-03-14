// Package main provides a simple hello world application.
package main

import (
	"log/slog"
	"os"
)

func main() {
	if err := run(); err != nil {
		slog.Error("run", "error", err)
		os.Exit(1)
	}
}

func run() error {
	slog.Info(Message())
	return nil
}

func Message() string {
	return "Hello World"
}
