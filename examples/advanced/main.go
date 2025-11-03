package main

import (
	"fmt"
	"log"
	"os"

	"github.com/alicanli1995/conform"
)

// AdvancedConfig demonstrates advanced features
type AdvancedConfig struct {
	// Custom validation with multiple rules
	APIKey string `conform:"env=API_KEY,required,validate=min:32,alphanum"`

	// One-of validation
	Mode string `conform:"env=MODE,default=development,validate=oneof:development:staging:production"`

	// Regex validation
	Version string `conform:"env=VERSION,default=1.0.0,validate=regex:^\\d+\\.\\d+\\.\\d+$"`

	// Hostname validation
	Domain string `conform:"env=DOMAIN,default=example.com,validate=hostname"`

	// IP validation
	BindIP string `conform:"env=BIND_IP,default=0.0.0.0,validate=ip"`

	// Length validation
	Secret string `conform:"env=SECRET,required,validate=len:64"`
}

func main() {
	os.Setenv("API_KEY", "abcdefghijklmnopqrstuvwxyz123456")
	os.Setenv("MODE", "production")
	os.Setenv("VERSION", "2.1.0")
	os.Setenv("DOMAIN", "api.example.com")
	os.Setenv("BIND_IP", "127.0.0.1")
	os.Setenv("SECRET", "a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6q7r8s9t0u1v2w3x4y5z6a7b8c9d0e1f2")

	var cfg AdvancedConfig
	if err := conform.Load(&cfg); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("API Key: %s\n", cfg.APIKey)
	fmt.Printf("Mode: %s\n", cfg.Mode)
	fmt.Printf("Version: %s\n", cfg.Version)
	fmt.Printf("Domain: %s\n", cfg.Domain)
	fmt.Printf("Bind IP: %s\n", cfg.BindIP)
	fmt.Printf("Secret: %s\n", cfg.Secret)
}
