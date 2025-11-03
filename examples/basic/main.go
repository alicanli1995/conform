package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/alicanli1995/conform"
)

// BasicConfig demonstrates basic usage with environment variables
type BasicConfig struct {
	Port     int    `conform:"env=PORT,default=8080,validate=gte:1024,lte:65535"`
	Host     string `conform:"env=HOST,default=localhost"`
	Debug    bool   `conform:"env=DEBUG,default=false"`
	Timeout  int    `conform:"env=TIMEOUT,default=30"`
	Database string `conform:"env=DATABASE_URL,required"`
}

func ExampleBasic() {
	// Set some environment variables for demonstration
	os.Setenv("PORT", "3000")
	os.Setenv("DATABASE_URL", "postgres://localhost/mydb")
	os.Setenv("DEBUG", "true")

	var cfg BasicConfig
	if err := conform.Load(&cfg); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Port: %d\n", cfg.Port)
	fmt.Printf("Host: %s\n", cfg.Host)
	fmt.Printf("Debug: %t\n", cfg.Debug)
	fmt.Printf("Timeout: %d\n", cfg.Timeout)
	fmt.Printf("Database: %s\n", cfg.Database)
}

// NestedConfig demonstrates nested structs with prefixes
type DatabaseConfig struct {
	Host     string `conform:"env=HOST,default=localhost"`
	Port     int    `conform:"env=PORT,default=5432"`
	User     string `conform:"env=USER,default=postgres"`
	Password string `conform:"env=PASSWORD,required"`
	DBName   string `conform:"env=DBNAME,default=mydb"`
}

type RedisConfig struct {
	Host     string `conform:"env=HOST,default=localhost"`
	Port     int    `conform:"env=PORT,default=6379"`
	Password string `conform:"env=PASSWORD"`
}

type AppConfig struct {
	Name     string         `conform:"env=APP_NAME,default=MyApp"`
	Database DatabaseConfig `conform:"prefix=DB_"`
	Redis    RedisConfig    `conform:"prefix=REDIS_"`
}

func ExampleNested() {
	// Set environment variables with prefixes
	os.Setenv("DB_HOST", "db.example.com")
	os.Setenv("DB_PORT", "5432")
	os.Setenv("DB_PASSWORD", "secret123")
	os.Setenv("REDIS_HOST", "redis.example.com")
	os.Setenv("REDIS_PORT", "6380")

	var cfg AppConfig
	if err := conform.Load(&cfg); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("App Name: %s\n", cfg.Name)
	fmt.Printf("Database: %s:%d\n", cfg.Database.Host, cfg.Database.Port)
	fmt.Printf("Redis: %s:%d\n", cfg.Redis.Host, cfg.Redis.Port)
}

// ValidationConfig demonstrates various validation rules
type ValidationConfig struct {
	Email    string `conform:"env=EMAIL,required,validate=email"`
	Port     int    `conform:"env=PORT,default=8080,validate=gte:1024,lte:65535"`
	Password string `conform:"env=PASSWORD,required,validate=min:8,has_upper,has_lower,has_digit"`
	URL      string `conform:"env=URL,required,validate=url:https"`
	Age      int    `conform:"env=AGE,default=18,validate=gte:18,lte:120"`
}

func ExampleValidation() {
	os.Setenv("EMAIL", "user@example.com")
	os.Setenv("PORT", "3000")
	os.Setenv("PASSWORD", "SecurePass123")
	os.Setenv("URL", "https://example.com")
	os.Setenv("AGE", "25")

	var cfg ValidationConfig
	if err := conform.Load(&cfg); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Email: %s\n", cfg.Email)
	fmt.Printf("Port: %d\n", cfg.Port)
	fmt.Printf("Password: %s\n", cfg.Password)
	fmt.Printf("URL: %s\n", cfg.URL)
	fmt.Printf("Age: %d\n", cfg.Age)
}

// SliceConfig demonstrates slice and array handling
type SliceConfig struct {
	Hosts    []string `conform:"env=HOSTS,default=localhost,separator=|"`
	Ports    []int    `conform:"env=PORTS,default=8080,separator=,"`
	Features []string `conform:"env=FEATURES,separator=,"`
}

func ExampleSlices() {
	os.Setenv("HOSTS", "host1|host2|host3")
	os.Setenv("PORTS", "8080,8081,8082")
	os.Setenv("FEATURES", "feature1,feature2,feature3")

	var cfg SliceConfig
	if err := conform.Load(&cfg); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Hosts: %v\n", cfg.Hosts)
	fmt.Printf("Ports: %v\n", cfg.Ports)
	fmt.Printf("Features: %v\n", cfg.Features)
}

// TimeConfig demonstrates time parsing
type TimeConfig struct {
	StartTime time.Time `conform:"env=START_TIME,default=2024-01-01T00:00:00Z,format=2006-01-02T15:04:05Z"`
	BirthDate time.Time `conform:"env=BIRTH_DATE,format=2006-01-02"`
}

func ExampleTime() {
	os.Setenv("START_TIME", "2024-01-01T00:00:00Z")
	os.Setenv("BIRTH_DATE", "1990-05-15")

	var cfg TimeConfig
	if err := conform.Load(&cfg); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Start Time: %s\n", cfg.StartTime.Format(time.RFC3339))
	fmt.Printf("Birth Date: %s\n", cfg.BirthDate.Format("2006-01-02"))
}

// FileConfig demonstrates loading from YAML files
type FileConfig struct {
	Server struct {
		Host string `conform:"file=server.host,default=localhost"`
		Port int    `conform:"file=server.port,default=8080"`
	}
	Database struct {
		Host string `conform:"file=database.host,default=localhost"`
		Port int    `conform:"file=database.port,default=5432"`
	}
}

func ExampleFileConfig() {
	// Create a sample config file
	configYAML := `server:
  host: api.example.com
  port: 3000
database:
  host: db.example.com
  port: 5432
`

	// Write to temp file
	tmpFile := "config.yaml"
	if err := os.WriteFile(tmpFile, []byte(configYAML), 0644); err != nil {
		log.Fatal(err)
	}
	defer os.Remove(tmpFile)

	var cfg FileConfig
	if err := conform.Load(&cfg, conform.WithFile(tmpFile)); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Server: %s:%d\n", cfg.Server.Host, cfg.Server.Port)
	fmt.Printf("Database: %s:%d\n", cfg.Database.Host, cfg.Database.Port)
}

// CombinedConfig demonstrates combining env vars and file config
type CombinedConfig struct {
	// From environment
	Environment string `conform:"env=ENV,default=development"`
	Port        int    `conform:"env=PORT,default=8080"`

	// From file
	API struct {
		Key string `conform:"file=api.key,required"`
		URL string `conform:"file=api.url,default=https://api.example.com"`
	}
}

func ExampleCombined() {
	configYAML := `api:
  key: secret-api-key-12345
  url: https://api.production.com
`

	tmpFile := "config.yaml"
	if err := os.WriteFile(tmpFile, []byte(configYAML), 0644); err != nil {
		log.Fatal(err)
	}
	defer os.Remove(tmpFile)

	os.Setenv("ENV", "production")
	os.Setenv("PORT", "3000")

	var cfg CombinedConfig
	if err := conform.Load(&cfg, conform.WithFile(tmpFile)); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Environment: %s\n", cfg.Environment)
	fmt.Printf("Port: %d\n", cfg.Port)
	fmt.Printf("API Key: %s\n", cfg.API.Key)
	fmt.Printf("API URL: %s\n", cfg.API.URL)
}

func main() {
	fmt.Println("=== Basic Config ===")
	ExampleBasic()
	fmt.Println("\n=== Nested Config ===")
	ExampleNested()
	fmt.Println("\n=== Validation Config ===")
	ExampleValidation()
	fmt.Println("\n=== Slice Config ===")
	ExampleSlices()
	fmt.Println("\n=== Time Config ===")
	ExampleTime()
	fmt.Println("\n=== File Config ===")
	ExampleFileConfig()
	fmt.Println("\n=== Combined Config ===")
	ExampleCombined()
}
