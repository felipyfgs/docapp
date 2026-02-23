package config

import (
	"docapp/core/internal/db"

	"github.com/rs/zerolog"
	"github.com/uptrace/bun"
)

func ConnectDB(dsn string, log zerolog.Logger) (*bun.DB, error) {
	return db.Connect(dsn, log)
}
