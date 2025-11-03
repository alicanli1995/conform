package main

import (
	"fmt"
	"log"
	"os"

	"github.com/alicanli1995/conform"
)

// RealWorldConfig demonstrates a realistic production configuration
type RealWorldConfig struct {
	App struct {
		Name        string `conform:"env=APP_NAME,default=MyApp"`
		Environment string `conform:"env=ENV,default=development,validate=oneof:development:staging:production"`
		Version     string `conform:"env=VERSION,default=1.0.0"`
		Debug       bool   `conform:"env=DEBUG,default=false"`
	}

	Server struct {
		Host         string `conform:"env=HOST,default=0.0.0.0"`
		Port         int    `conform:"env=PORT,default=8080,validate=gte:1024,lte:65535"`
		ReadTimeout  int    `conform:"env=READ_TIMEOUT,default=30"`
		WriteTimeout int    `conform:"env=WRITE_TIMEOUT,default=30"`
	}

	Database struct {
		Host     string `conform:"env=DB_HOST,default=localhost"`
		Port     int    `conform:"env=DB_PORT,default=5432"`
		User     string `conform:"env=DB_USER,default=postgres"`
		Password string `conform:"env=DB_PASSWORD,required"`
		DBName   string `conform:"env=DB_NAME,default=mydb"`
		SSLMode  string `conform:"env=DB_SSL_MODE,default=disable,validate=oneof:disable:require:verify-ca:verify-full"`
		MaxConns int    `conform:"env=DB_MAX_CONNS,default=25"`
	}

	Redis struct {
		Host     string `conform:"env=REDIS_HOST,default=localhost"`
		Port     int    `conform:"env=REDIS_PORT,default=6379"`
		Password string `conform:"env=REDIS_PASSWORD"`
		DB       int    `conform:"env=REDIS_DB,default=0"`
	}

	JWT struct {
		Secret     string `conform:"env=JWT_SECRET,required,validate=min:32"`
		Expiration int    `conform:"env=JWT_EXPIRATION,default=3600"`
	}

	API struct {
		Key string `conform:"env=API_KEY,required,validate=min:32"`
		URL string `conform:"env=API_URL,required,validate=url:https"`
	}
}

func main() {
	// Set environment variables
	os.Setenv("APP_NAME", "ProductionApp")
	os.Setenv("ENV", "production")
	os.Setenv("PORT", "3000")
	os.Setenv("DB_HOST", "db.example.com")
	os.Setenv("DB_PASSWORD", "secure_password_123")
	os.Setenv("REDIS_HOST", "redis.example.com")
	os.Setenv("JWT_SECRET", "my-super-secret-jwt-key-that-is-at-least-32-chars")
	os.Setenv("API_KEY", "api-key-that-is-at-least-32-characters-long")
	os.Setenv("API_URL", "https://api.example.com")

	var cfg RealWorldConfig
	if err := conform.Load(&cfg); err != nil {
		log.Fatal(err)
	}

	fmt.Println("=== Application Configuration ===")
	fmt.Printf("Name: %s\n", cfg.App.Name)
	fmt.Printf("Environment: %s\n", cfg.App.Environment)
	fmt.Printf("Version: %s\n", cfg.App.Version)
	fmt.Printf("Debug: %t\n", cfg.App.Debug)

	fmt.Println("\n=== Server Configuration ===")
	fmt.Printf("Host: %s\n", cfg.Server.Host)
	fmt.Printf("Port: %d\n", cfg.Server.Port)
	fmt.Printf("Read Timeout: %d\n", cfg.Server.ReadTimeout)
	fmt.Printf("Write Timeout: %d\n", cfg.Server.WriteTimeout)

	fmt.Println("\n=== Database Configuration ===")
	fmt.Printf("Host: %s\n", cfg.Database.Host)
	fmt.Printf("Port: %d\n", cfg.Database.Port)
	fmt.Printf("User: %s\n", cfg.Database.User)
	fmt.Printf("Password: %s\n", maskPassword(cfg.Database.Password))
	fmt.Printf("DB Name: %s\n", cfg.Database.DBName)
	fmt.Printf("SSL Mode: %s\n", cfg.Database.SSLMode)
	fmt.Printf("Max Connections: %d\n", cfg.Database.MaxConns)

	fmt.Println("\n=== Redis Configuration ===")
	fmt.Printf("Host: %s\n", cfg.Redis.Host)
	fmt.Printf("Port: %d\n", cfg.Redis.Port)
	if cfg.Redis.Password != "" {
		fmt.Printf("Password: %s\n", maskPassword(cfg.Redis.Password))
	}
	fmt.Printf("DB: %d\n", cfg.Redis.DB)

	fmt.Println("\n=== JWT Configuration ===")
	fmt.Printf("Secret: %s\n", maskPassword(cfg.JWT.Secret))
	fmt.Printf("Expiration: %d seconds\n", cfg.JWT.Expiration)

	fmt.Println("\n=== API Configuration ===")
	fmt.Printf("Key: %s\n", maskPassword(cfg.API.Key))
	fmt.Printf("URL: %s\n", cfg.API.URL)
}

func maskPassword(pwd string) string {
	if len(pwd) <= 4 {
		return "****"
	}
	return pwd[:2] + "****" + pwd[len(pwd)-2:]
}
