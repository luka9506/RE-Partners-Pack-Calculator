package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
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

func Save(path string, cfg AppConfig) error {
	copyCfg := AppConfig{PackSizes: append([]int(nil), cfg.PackSizes...)}
	if err := copyCfg.Validate(); err != nil {
		return err
	}

	payload, err := json.MarshalIndent(copyCfg, "", "  ")
	if err != nil {
		return fmt.Errorf("encode config: %w", err)
	}
	payload = append(payload, '\n')

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create config directory: %w", err)
	}

	tempFile, err := os.CreateTemp(filepath.Dir(path), "packs-*.json")
	if err != nil {
		return fmt.Errorf("create temp config: %w", err)
	}
	tempPath := tempFile.Name()

	if _, err := tempFile.Write(payload); err != nil {
		_ = tempFile.Close()
		_ = os.Remove(tempPath)
		return fmt.Errorf("write config: %w", err)
	}

	if err := tempFile.Close(); err != nil {
		_ = os.Remove(tempPath)
		return fmt.Errorf("close temp config: %w", err)
	}

	if err := os.Rename(tempPath, path); err != nil {
		_ = os.Remove(tempPath)
		return fmt.Errorf("replace config: %w", err)
	}

	return nil
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
