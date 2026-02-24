package config

import (
	"os"
	"strconv"
)

type Config struct {
	Port                  string
	Env                   string
	SpedServiceURL        string
	SpedTimeoutSeconds    int
	DatabaseURL           string
	WorkerIntervalMinutes int
	CertsDir              string
	StorageEndpoint       string
	StorageAccessKey      string
	StorageSecretKey      string
	StorageBucket         string
	StorageUseSSL         bool
	StorageRegion         string
	StoragePresignMinutes int
	ADNBaseURL            string
}

func Load() *Config {
	return &Config{
		Port:                  getEnv("PORT", "8080"),
		Env:                   getEnv("ENV", "development"),
		SpedServiceURL:        getEnv("SPED_SERVICE_URL", "http://localhost:8000/api"),
		SpedTimeoutSeconds:    getEnvInt("SPED_TIMEOUT_SECONDS", 180),
		DatabaseURL:           getEnv("DATABASE_URL", "postgres://fiscal:fiscal@localhost:5432/fiscal?sslmode=disable"),
		WorkerIntervalMinutes: getEnvInt("WORKER_INTERVAL_MINUTES", 30),
		CertsDir:              getEnv("CERTS_DIR", "certs"),
		StorageEndpoint:       getEnv("STORAGE_ENDPOINT", "localhost:9000"),
		StorageAccessKey:      getEnv("STORAGE_ACCESS_KEY", "minioadmin"),
		StorageSecretKey:      getEnv("STORAGE_SECRET_KEY", "minioadmin"),
		StorageBucket:         getEnv("STORAGE_BUCKET", "documentos"),
		StorageUseSSL:         getEnvBool("STORAGE_USE_SSL", false),
		StorageRegion:         getEnv("STORAGE_REGION", "us-east-1"),
		StoragePresignMinutes: getEnvInt("STORAGE_PRESIGN_MINUTES", 15),
		ADNBaseURL:            getEnv("ADN_BASE_URL", "https://adn.nfse.gov.br/contribuintes"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}

	return parsed
}

func getEnvBool(key string, fallback bool) bool {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}

	parsed, err := strconv.ParseBool(v)
	if err != nil {
		return fallback
	}

	return parsed
}
