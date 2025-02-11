package store

import (
	"fmt"
	"log"

	"github.com/SobolevTim/t-graphql/internal/config"
)

// NewStore создаёт нужную реализацию хранилища
func NewStore(cfg *config.Config) (Store, error) {
	switch cfg.StorageType {
	case "postgres":
		store, err := NewPostgresStore(cfg.DatabaseURL)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to Postgres: %w", err)
		}
		log.Println("Using PostgreSQL storage")
		return store, nil
	case "memory":
		log.Println("Using in-memory storage")
		return NewMemoryStore(), nil
	default:
		return nil, fmt.Errorf("unknown storage type: %s", cfg.StorageType)
	}
}
