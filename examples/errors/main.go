package main

import (
	"fmt"
	"os"

	"github.com/alicanli1995/conform"
)

type Config struct {
	Port     int            `conform:"env=APP_PORT,default=8080,validate=gte:1024"`
	Host     string         `conform:"env=APP_HOST,default=localhost,validate=hostname"`
	Database DatabaseConfig `conform:"prefix=DB_"`
}

type DatabaseConfig struct {
	URL      string `conform:"env=URL,required,validate=url"`
	MaxConns int    `conform:"env=MAX_CONNS,default=10,validate=gte:1,lte:100"`
}

func main() {
	// Demonstrate beautiful error messages
	os.Setenv("APP_PORT", "80")             // Too small - will fail validation
	os.Setenv("APP_HOST", "not-a-hostname") // Invalid hostname
	// Missing DB_URL - required field

	_, err := conform.LoadGeneric[Config](conform.FromEnv())
	if err != nil {
		fmt.Println(err)
		// Output:
		// ❌ Configuration validation failed:
		//
		// 1. Port (APP_PORT): validation 'gte' failed: value 80 is less than 1024
		//    Got: 80
		//    Location: env var APP_PORT
		//    💡 Suggestion: Use a value >= 1024
		//
		// 2. Host (APP_HOST): validation 'hostname' failed: invalid hostname
		//    Got: not-a-hostname
		//    Location: env var APP_HOST
		//    💡 Suggestion: Format should be: example.com
		//
		// 3. Database.URL (DB_URL): missing required field
		//    💡 Suggestion: Set via: DB_URL environment variable
	}
}
