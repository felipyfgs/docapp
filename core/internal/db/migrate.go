package db

import (
	"embed"
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/rs/zerolog"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

func RunMigrations(dsn string, log zerolog.Logger) error {
	sourceDriver, err := iofs.New(migrationsFS, "migrations")
	if err != nil {
		return fmt.Errorf("creating migration source: %w", err)
	}

	m, err := migrate.NewWithSourceInstance("iofs", sourceDriver, dsn)
	if err != nil {
		return fmt.Errorf("creating migrate instance: %w", err)
	}

	defer func() {
		_, _ = m.Close()
	}()

	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			log.Debug().Msg("database migrations already up-to-date")
			return nil
		}
		return fmt.Errorf("running migrations: %w", err)
	}

	log.Info().Msg("database migrations applied")
	return nil
}
