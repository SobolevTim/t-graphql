package config_test

import (
	"os"
	"testing"

	"github.com/SobolevTim/t-graphql/internal/config"
)

func TestLoadConfig_DefaultValues(t *testing.T) {
	// Unset environment variables to ensure defaults are used
	os.Unsetenv("STORAGE_TYPE")
	os.Unsetenv("DATABASE_URL")

	config := config.LoadConfig()

	if config.StorageType != "memory" {
		t.Errorf("Expected StorageType to be 'memory', got '%s'", config.StorageType)
	}

	expectedDatabaseURL := "postgres://user:pass@localhost:5432/db"
	if config.DatabaseURL != expectedDatabaseURL {
		t.Errorf("Expected DatabaseURL to be '%s', got '%s'", expectedDatabaseURL, config.DatabaseURL)
	}
}

func TestLoadConfig_WithEnvVariables(t *testing.T) {
	// Set environment variables
	os.Setenv("STORAGE_TYPE", "disk")
	os.Setenv("DATABASE_URL", "postgres://newuser:newpass@localhost:5432/newdb")

	config := config.LoadConfig()

	if config.StorageType != "disk" {
		t.Errorf("Expected StorageType to be 'disk', got '%s'", config.StorageType)
	}

	expectedDatabaseURL := "postgres://newuser:newpass@localhost:5432/newdb"
	if config.DatabaseURL != expectedDatabaseURL {
		t.Errorf("Expected DatabaseURL to be '%s', got '%s'", expectedDatabaseURL, config.DatabaseURL)
	}

	// Clean up environment variables
	os.Unsetenv("STORAGE_TYPE")
	os.Unsetenv("DATABASE_URL")
}
