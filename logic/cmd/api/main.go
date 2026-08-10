package main

import (
	"os"

	"github.com/english/backend/internal/bootstrap"
	"github.com/rs/zerolog/log"
)

func main() {
	if err := bootstrap.Start(); err != nil {
		log.Error().Err(err).Msg("application error")
		os.Exit(1)
	}
}
