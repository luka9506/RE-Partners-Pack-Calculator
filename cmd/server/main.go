package main

import (
	"log"
	"net/http"
	"os"

	"github.com/example/repartners-pack-calculator/internal/config"
	"github.com/example/repartners-pack-calculator/internal/httpapi"
	"github.com/example/repartners-pack-calculator/internal/packing"
)

func main() {
	// Keep operational settings external so the same binary works locally,
	// in containers, and on hosted platforms without rebuilds.
	configPath := envOrDefault("PACK_CONFIG_PATH", "config/packs.json")
	listenAddr := envOrDefault("PORT", "8080")

	cfg, err := config.Load(configPath)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	calculator, err := packing.NewCalculator(cfg.PackSizes)
	if err != nil {
		log.Fatalf("create calculator: %v", err)
	}

	server, err := httpapi.NewServer(calculator, configPath, cfg.PackSizes)
	if err != nil {
		log.Fatalf("create server: %v", err)
	}

	log.Printf("listening on :%s", listenAddr)
	if err := http.ListenAndServe(":"+listenAddr, server.Routes()); err != nil {
		log.Fatalf("serve: %v", err)
	}
}

func envOrDefault(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
