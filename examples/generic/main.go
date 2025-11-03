package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/alicanli1995/conform"
)

// Config demonstrates the new generic API
type Config struct {
	Port     int            `conform:"env=APP_PORT,default=8080,validate=gte:1024"`
	Host     string         `conform:"env=APP_HOST,default=localhost,validate=hostname"`
	Debug    bool           `conform:"env=DEBUG,default=false"`
	Timeout  time.Duration  `conform:"env=TIMEOUT,default=30s"`
	Database DatabaseConfig `conform:"prefix=DB_"`
}

type DatabaseConfig struct {
	URL      string `conform:"env=URL,required,validate=url"`
	MaxConns int    `conform:"env=MAX_CONNS,default=10,validate=gte:1,lte:100"`
}

func main() {
	// Set environment variables
	os.Setenv("APP_PORT", "3000")
	os.Setenv("APP_HOST", "api.example.com")
	os.Setenv("DEBUG", "true")
	os.Setenv("TIMEOUT", "60s")
	os.Setenv("DB_URL", "postgres://user:pass@localhost/db")
	os.Setenv("DB_MAX_CONNS", "25")

	// ✨ NEW: Generic API - tek satır!
	cfg, err := conform.LoadGeneric[Config](
		conform.FromEnv(),
		conform.FromFile("config.yaml"), // Optional
	)

	if err != nil {
		log.Fatal(err) // Beautiful error messages!
	}

	// cfg artık type-safe, validated, ready to use!
	fmt.Printf("Starting on %s:%d\n", cfg.Host, cfg.Port)
	fmt.Printf("Database: %s\n", cfg.Database.URL)
	fmt.Printf("Timeout: %s\n", cfg.Timeout)
	fmt.Printf("Debug: %t\n", cfg.Debug)
}
