package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Values map[string]string
}

// LoadConfig loads configuration from .config file in the project root
func LoadConfig() (*Config, error) {
	// Load from .config file in the project root
	err := godotenv.Load(".config")
	if err != nil {
		return nil, fmt.Errorf("error loading .config file: %w", err)
	}

	cfg := &Config{
		Values: make(map[string]string),
	}

	// Iterate over all environment variables loaded by godotenv
	for _, env := range os.Environ() {
		// env is in the form "KEY=VALUE"
		var key, value string
		n := 0
		for i, c := range env {
			if c == '=' {
				key = env[:i]
				value = env[i+1:]
				n = 1
				break
			}
		}
		if n == 1 {
			cfg.Values[key] = value
		}
	}

	return cfg, nil
}
