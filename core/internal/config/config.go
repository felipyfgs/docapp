package config

import (
	"os"
	"strconv"
)

type Config struct {
	Port               string
	Env                string
	SpedServiceURL     string
	SpedTimeoutSeconds int
}

func Load() *Config {
	return &Config{
		Port:               getEnv("PORT", "8080"),
		Env:                getEnv("ENV", "development"),
		SpedServiceURL:     getEnv("SPED_SERVICE_URL", "http://sped:8000"),
		SpedTimeoutSeconds: getEnvInt("SPED_TIMEOUT_SECONDS", 15),
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
