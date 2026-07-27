package main

import (
	"log/slog"
	"os"

	"github.com/english/backend/internal/bootstrap"
)

func main() {
	if err := bootstrap.Start(); err != nil {
		slog.Error("application error", "error", err)
		os.Exit(1)
	}
}
