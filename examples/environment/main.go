package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"

	"github.com/alicanli1995/conform"
)

// EnvironmentConfig demonstrates environment-specific configuration
type EnvironmentConfig struct {
	App struct {
		Name        string `conform:"env=APP_NAME,default=MyApp"`
		Environment string `conform:"env=ENV,default=development"`
	}
	Database struct {
		Host string `conform:"env=DB_HOST,file=database.host,default=localhost"`
		Port int    `conform:"env=DB_PORT,file=database.port,default=5432"`
		URL  string `conform:"file=database.url,default=postgres://localhost/mydb"`
	}
	Server struct {
		Host string `conform:"env=SERVER_HOST,file=server.host,default=0.0.0.0"`
		Port int    `conform:"env=SERVER_PORT,file=server.port,default=8080"`
	}
}

// getExampleDir returns the directory where this example is located
func getExampleDir() string {
	_, filename, _, _ := runtime.Caller(0)
	return filepath.Dir(filename)
}

func main() {
	exampleDir := getExampleDir()

	fmt.Println("=== Environment-Specific Configuration Demo ===")

	// Example 1: Development environment
	fmt.Println("📦 Loading DEVELOPMENT config...")
	devConfig, err := conform.LoadGeneric[EnvironmentConfig](
		conform.WithEnvironment("development"),
		conform.FromFile(filepath.Join(exampleDir, "config.${ENV}.yaml")),
		conform.FromEnv(),
	)
	if err != nil {
		log.Printf("Error loading dev config: %v", err)
	} else {
		fmt.Printf("✅ App: %s (Environment: %s)\n", devConfig.App.Name, devConfig.App.Environment)
		fmt.Printf("✅ Database: %s:%d\n", devConfig.Database.Host, devConfig.Database.Port)
		fmt.Printf("✅ Server: %s:%d\n", devConfig.Server.Host, devConfig.Server.Port)
	}

	// Example 2: Production environment
	fmt.Println("\n📦 Loading PRODUCTION config...")
	prodConfig, err := conform.LoadGeneric[EnvironmentConfig](
		conform.WithEnvironment("production"),
		conform.FromFile(filepath.Join(exampleDir, "config.${ENV}.yaml")),
		conform.FromEnv(),
	)
	if err != nil {
		log.Printf("Error loading prod config: %v", err)
	} else {
		fmt.Printf("✅ App: %s (Environment: %s)\n", prodConfig.App.Name, prodConfig.App.Environment)
		fmt.Printf("✅ Database: %s:%d\n", prodConfig.Database.Host, prodConfig.Database.Port)
		fmt.Printf("✅ Server: %s:%d\n", prodConfig.Server.Host, prodConfig.Server.Port)
	}

	// Example 3: Variable substitution in values
	fmt.Println("\n📦 Variable Substitution Demo...")
	type SubstitutionConfig struct {
		DatabaseURL string `conform:"env=DB_URL,default=postgres://${DB_USER:-postgres}:${DB_PASSWORD}@${DB_HOST:-localhost}:${DB_PORT:-5432}/${DB_NAME:-mydb}"`
		APIURL      string `conform:"env=API_URL,default=https://api.${ENV:-development}.example.com"`
	}

	os.Setenv("DB_USER", "admin")
	os.Setenv("DB_PASSWORD", "secret123")
	os.Setenv("DB_HOST", "db.prod.com")

	subConfig, err := conform.LoadGeneric[SubstitutionConfig](
		conform.WithEnvironment("production"),
		conform.FromEnv(),
	)
	if err != nil {
		log.Printf("Error: %v", err)
	} else {
		fmt.Printf("✅ Database URL: %s\n", subConfig.DatabaseURL)
		fmt.Printf("✅ API URL: %s\n", subConfig.APIURL)
	}

	os.Unsetenv("DB_USER")
	os.Unsetenv("DB_PASSWORD")
	os.Unsetenv("DB_HOST")
}
