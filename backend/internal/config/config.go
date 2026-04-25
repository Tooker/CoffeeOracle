package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config contains runtime configuration for the CoffeeOracle backend service.
type Config struct {
	Port         string
	OpenAIAPIKey string
	LogLevel     string
	LogEnabled   bool
}

// Addr returns the bind address in the form ":<port>".
func (c Config) Addr() string {
	return fmt.Sprintf(":%s", c.Port)
}

// Load reads configuration from the environment (optionally loading a local .env file)
// and validates all required fields.
func Load() (Config, error) {
	// Load .env if present; ignore error so deployments relying purely on env vars still work.
	_ = godotenv.Load()

	cfg := Config{
		Port:         os.Getenv("PORT"),
		OpenAIAPIKey: os.Getenv("OPENAI_API_KEY"),
		LogLevel:     getEnvWithDefault("LOG_LEVEL", "debug"),
		LogEnabled:   parseBoolWithDefault(os.Getenv("LOG_ENABLED"), true),
	}

	if cfg.Port == "" {
		return Config{}, errors.New("PORT is required")
	}
	if _, err := strconv.Atoi(cfg.Port); err != nil {
		return Config{}, fmt.Errorf("PORT must be numeric: %w", err)
	}

	if cfg.OpenAIAPIKey == "" {
		return Config{}, errors.New("OPENAI_API_KEY is required")
	}

	return cfg, nil
}

// getEnvWithDefault reads an environment variable and falls back to a default value.
// This keeps the app runnable even when optional settings are not provided.
func getEnvWithDefault(key, def string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return def
}

// parseBoolWithDefault converts text like "true"/"false" into a boolean.
// If parsing fails, we safely fall back to a known default.
func parseBoolWithDefault(val string, def bool) bool {
	if val == "" {
		return def
	}
	parsed, err := strconv.ParseBool(val)
	if err != nil {
		return def
	}
	return parsed
}
