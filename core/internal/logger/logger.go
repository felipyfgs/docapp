package logger

import (
	"os"
	"time"

	"github.com/rs/zerolog"
)

func New(env string) zerolog.Logger {
	zerolog.TimeFieldFormat = time.RFC3339

	if env == "development" {
		return zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: "15:04:05"}).
			With().Timestamp().Caller().Logger().
			Level(zerolog.DebugLevel)
	}

	return zerolog.New(os.Stderr).
		With().Timestamp().Logger().
		Level(zerolog.InfoLevel)
}
