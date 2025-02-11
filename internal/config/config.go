package config

import (
	"os"
)

// Config хранит параметры конфигурации
type Config struct {
	StorageType string
	DatabaseURL string
}

// LoadConfig загружает конфигурацию из переменных окружения
func LoadConfig() *Config {
	return &Config{
		StorageType: getEnv("STORAGE_TYPE", "memory"), // По умолчанию in-memory
		DatabaseURL: getEnv("DATABASE_URL", "postgres://user:pass@localhost:5432/db"),
	}
}

// getEnv получает значение переменной окружения или использует значение по умолчанию
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
