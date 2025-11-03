package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/alicanli1995/conform"
)

// Config demonstrates hot reload
type Config struct {
	Port     int       `conform:"env=PORT,default=8080"`
	Host     string    `conform:"env=HOST,default=localhost"`
	Debug    bool      `conform:"env=DEBUG,default=false"`
	Reloaded time.Time `conform:"-"`
}

func main() {
	// Hot reload example
	watcher, err := conform.Watch[Config](func(newCfg Config) {
		fmt.Printf("🔄 Config reloaded at %s\n", time.Now().Format(time.RFC3339))
		fmt.Printf("   Port: %d\n", newCfg.Port)
		fmt.Printf("   Host: %s\n", newCfg.Host)
		fmt.Printf("   Debug: %t\n", newCfg.Debug)
	}, conform.FromEnv())

	if err != nil {
		log.Fatal(err)
	}

	// Get initial config
	cfg := watcher.Get()
	fmt.Printf("Initial config: %+v\n", cfg)

	// Simulate config changes
	fmt.Println("\nChanging PORT to 3000...")
	os.Setenv("PORT", "3000")
	time.Sleep(2 * time.Second)

	fmt.Println("\nChanging DEBUG to true...")
	os.Setenv("DEBUG", "true")
	time.Sleep(2 * time.Second)

	// Get updated config
	cfg = watcher.Get()
	fmt.Printf("\nFinal config: %+v\n", cfg)

	// Stop watching
	watcher.Stop()
	fmt.Println("\nWatcher stopped")
}
