package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	SourceURL string
	TargetURL string
	AuthToken string
	Port      string
}

func Load() (*Config, error) {
	err := godotenv.Load()
	if err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("error loading .env file: %w", err)
	}

	cfg := &Config{
		SourceURL: os.Getenv("SOURCE_URL"),
		TargetURL: os.Getenv("TARGET_URL"),
		AuthToken: os.Getenv("AUTH_TOKEN"),
		Port:      os.Getenv("PORT"),
	}

	if cfg.SourceURL == "" {
		return nil, fmt.Errorf("SOURCE_URL is required")
	}
	if cfg.TargetURL == "" {
		return nil, fmt.Errorf("TARGET_URL is required")
	}
	if cfg.Port == "" {
		cfg.Port = "8080"
	}

	return cfg, nil
}
