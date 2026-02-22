package config

import (
	"docapp/core/internal/model"

	"github.com/rs/zerolog"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

func ConnectDB(dsn string, log zerolog.Logger) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Warn),
	})
	if err != nil {
		log.Error().Err(err).Msg("database connection failed")
		return nil, err
	}

	if err := db.AutoMigrate(&model.Empresa{}, &model.DocumentoFiscal{}); err != nil {
		log.Error().Err(err).Msg("database migration failed")
		return nil, err
	}

	log.Info().Msg("database connected and migrated")

	return db, nil
}
