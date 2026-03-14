// Package main demonstrates third-party library usage.
package main

import (
	"log/slog"

	"github.com/google/uuid"
)

func main() {
	uuidString := uuid.NewString()
	slog.Info("UUID generated", "uuid", uuidString)
}
