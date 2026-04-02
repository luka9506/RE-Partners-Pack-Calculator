package config

import (
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestLoadNormalizesPackSizes(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "packs.json")
	if err := Save(configPath, AppConfig{PackSizes: []int{1000, 250, 500}}); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	expected := []int{250, 500, 1000}
	if !reflect.DeepEqual(cfg.PackSizes, expected) {
		t.Fatalf("PackSizes = %#v, want %#v", cfg.PackSizes, expected)
	}
}

func TestSaveRejectsInvalidPackSizes(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "packs.json")

	err := Save(configPath, AppConfig{PackSizes: []int{250, 250}})
	if err == nil {
		t.Fatal("Save() expected duplicate pack size error")
	}
}

func TestLoadRejectsInvalidConfig(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "packs.json")
	if err := Save(configPath, AppConfig{PackSizes: []int{250, 500}}); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	err := Save(configPath, AppConfig{PackSizes: []int{0, 500}})
	if err == nil {
		t.Fatal("Save() expected error for zero pack size")
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	expected := []int{250, 500}
	if !reflect.DeepEqual(cfg.PackSizes, expected) {
		t.Fatalf("PackSizes = %#v, want %#v", cfg.PackSizes, expected)
	}
}

func TestValidateErrors(t *testing.T) {
	tests := []struct {
		name string
		cfg  AppConfig
	}{
		{name: "empty", cfg: AppConfig{}},
		{name: "negative", cfg: AppConfig{PackSizes: []int{-1, 250}}},
		{name: "duplicate", cfg: AppConfig{PackSizes: []int{250, 250}}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := test.cfg.Validate(); err == nil {
				t.Fatal("Validate() expected error")
			}
		})
	}
}

func TestLoadMissingFile(t *testing.T) {
	_, err := Load(filepath.Join(t.TempDir(), "missing.json"))
	if err == nil {
		t.Fatal("Load() expected error for missing file")
	}
	if !strings.Contains(err.Error(), "open config") {
		t.Fatalf("error = %v, want open config context", err)
	}
}
