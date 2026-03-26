package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
)

type AppConfig struct {
	PackSizes []int `json:"pack_sizes"`
}

func Load(path string) (AppConfig, error) {
	file, err := os.Open(path)
	if err != nil {
		return AppConfig{}, fmt.Errorf("open config: %w", err)
	}
	defer file.Close()

	var cfg AppConfig
	if err := json.NewDecoder(file).Decode(&cfg); err != nil {
		return AppConfig{}, fmt.Errorf("decode config: %w", err)
	}

	// Validation also sorts the pack sizes so the rest of the application can use
	// the configuration without repeating normalization logic.
	if err := cfg.Validate(); err != nil {
		return AppConfig{}, err
	}

	return cfg, nil
}

func (c *AppConfig) Validate() error {
	if len(c.PackSizes) == 0 {
		return errors.New("config must include at least one pack size")
	}

	// Reject malformed configuration early so the service fails fast on startup.
	seen := make(map[int]struct{}, len(c.PackSizes))
	for _, size := range c.PackSizes {
		if size <= 0 {
			return fmt.Errorf("pack sizes must be positive: %d", size)
		}
		if _, exists := seen[size]; exists {
			return fmt.Errorf("duplicate pack size: %d", size)
		}
		seen[size] = struct{}{}
	}

	sort.Ints(c.PackSizes)
	return nil
}
