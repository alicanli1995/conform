package main

import (
	"fmt"
	"log"
	"path/filepath"
	"runtime"

	"github.com/alicanli1995/conform"
)

// FileConfig demonstrates loading configuration from files (YAML, JSON, TOML)
type FileConfig struct {
	Server struct {
		Host string `conform:"file=server.host,default=localhost"`
		Port int    `conform:"file=server.port,default=8080"`
	}
	Database struct {
		URL      string `conform:"file=database.url,required"`
		MaxConns int    `conform:"file=database.max_conns,default=10"`
	}
	Features struct {
		CacheEnabled bool   `conform:"file=features.cache_enabled,default=false"`
		LogLevel     string `conform:"file=features.log_level,default=info"`
	}
}

// getExampleDir returns the directory where this example is located
func getExampleDir() string {
	// Try to get the source file location
	_, filename, _, _ := runtime.Caller(0)
	return filepath.Dir(filename)
}

func main() {
	// Get the directory where this example is located
	exampleDir := getExampleDir()

	// Example 1: Load from YAML file
	fmt.Println("=== Loading from YAML ===")
	yamlPath := filepath.Join(exampleDir, "config.yaml")
	fmt.Printf("Looking for config at: %s\n", yamlPath)
	cfg, err := conform.LoadGeneric[FileConfig](conform.FromFile(yamlPath))
	if err != nil {
		log.Printf("Error loading YAML: %v", err)
	} else {
		fmt.Printf("✅ Server: %s:%d\n", cfg.Server.Host, cfg.Server.Port)
		fmt.Printf("✅ Database: %s (max connections: %d)\n", cfg.Database.URL, cfg.Database.MaxConns)
		fmt.Printf("✅ Features: Cache=%v, LogLevel=%s\n", cfg.Features.CacheEnabled, cfg.Features.LogLevel)
	}

	// Example 2: Load from TOML file
	fmt.Println("\n=== Loading from TOML ===")
	tomlPath := filepath.Join(exampleDir, "config.toml")
	fmt.Printf("Looking for config at: %s\n", tomlPath)
	cfg, err = conform.LoadGeneric[FileConfig](conform.FromFile(tomlPath))
	if err != nil {
		log.Printf("Error loading TOML: %v", err)
	} else {
		fmt.Printf("✅ Server: %s:%d\n", cfg.Server.Host, cfg.Server.Port)
		fmt.Printf("✅ Database: %s (max connections: %d)\n", cfg.Database.URL, cfg.Database.MaxConns)
		fmt.Printf("✅ Features: Cache=%v, LogLevel=%s\n", cfg.Features.CacheEnabled, cfg.Features.LogLevel)
	}

	// Example 3: Load from JSON file
	fmt.Println("\n=== Loading from JSON ===")
	jsonPath := filepath.Join(exampleDir, "config.json")
	fmt.Printf("Looking for config at: %s\n", jsonPath)
	cfg, err = conform.LoadGeneric[FileConfig](conform.FromFile(jsonPath))
	if err != nil {
		log.Printf("Error loading JSON: %v", err)
	} else {
		fmt.Printf("✅ Server: %s:%d\n", cfg.Server.Host, cfg.Server.Port)
		fmt.Printf("✅ Database: %s (max connections: %d)\n", cfg.Database.URL, cfg.Database.MaxConns)
		fmt.Printf("✅ Features: Cache=%v, LogLevel=%s\n", cfg.Features.CacheEnabled, cfg.Features.LogLevel)
	}

	// Example 4: Multi-source (File + Environment)
	fmt.Println("\n=== Multi-source (File + Env) ===")
	type MultiConfig struct {
		Server struct {
			Host string `conform:"env=SERVER_HOST,file=server.host,default=localhost"`
			Port int    `conform:"env=SERVER_PORT,file=server.port,default=8080"`
		}
	}

	cfg2, err := conform.LoadGeneric[MultiConfig](
		conform.FromEnv(),
		conform.FromFile(yamlPath),
	)
	if err != nil {
		log.Printf("Error: %v", err)
	} else {
		fmt.Printf("✅ Server: %s:%d (env vars take priority)\n", cfg2.Server.Host, cfg2.Server.Port)
	}
}
