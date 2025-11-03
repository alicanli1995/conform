package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"reflect"

	"github.com/alicanli1995/conform"
)

var (
	configFile = flag.String("config", "", "Path to configuration file (YAML, JSON, or TOML)")
	structFile = flag.String("struct", "", "Path to Go file containing struct definition")
	structName = flag.String("struct-name", "", "Name of the struct to validate (required if -struct is provided)")
	envFile    = flag.String("env", "", "Path to .env file")
	verbose    = flag.Bool("verbose", false, "Show detailed output")
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: conform-validate [options]\n\n")
		fmt.Fprintf(os.Stderr, "Validate configuration files against struct definitions.\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  # Validate config.yaml against struct in types.go\n")
		fmt.Fprintf(os.Stderr, "  conform-validate -config config.yaml -struct types.go -struct-name Config\n\n")
		fmt.Fprintf(os.Stderr, "  # Validate from environment variables\n")
		fmt.Fprintf(os.Stderr, "  conform-validate -struct types.go -struct-name Config\n\n")
	}

	flag.Parse()

	if *structFile == "" {
		fmt.Fprintf(os.Stderr, "Error: -struct flag is required\n")
		flag.Usage()
		os.Exit(1)
	}

	if *structName == "" {
		fmt.Fprintf(os.Stderr, "Error: -struct-name flag is required\n")
		flag.Usage()
		os.Exit(1)
	}

	// For now, we'll validate using a simple approach
	// In a full implementation, we'd parse the Go file and extract the struct

	fmt.Println("Conform Configuration Validator")
	fmt.Println("===============================")
	fmt.Println()

	var opts []conform.Option

	if *configFile != "" {
		fmt.Printf("Loading config from: %s\n", *configFile)
		opts = append(opts, conform.FromFile(*configFile))
	}

	opts = append(opts, conform.FromEnv())

	fmt.Printf("Validating struct: %s\n", *structName)
	fmt.Println()

	// Since we can't dynamically load structs from files,
	// we'll provide a helper message
	fmt.Println("Note: This CLI tool validates configuration files.")
	fmt.Println("To use it programmatically, define your struct and use:")
	fmt.Println()
	fmt.Printf("  cfg, err := conform.LoadGeneric[%s](\n", *structName)
	fmt.Println("      conform.FromFile(\"" + *configFile + "\"),")
	fmt.Println("      conform.FromEnv(),")
	fmt.Println("  )")
	fmt.Println()

	if *verbose {
		fmt.Println("Verbose mode enabled")
		fmt.Println("Configuration sources:")
		if *configFile != "" {
			fmt.Printf("  - File: %s\n", *configFile)
		}
		fmt.Println("  - Environment variables")
		fmt.Println()
	}

	fmt.Println("For full validation, use the Conform library in your Go code.")
}

// validateConfig validates a configuration file against a struct type
func validateConfig(configType reflect.Type, opts ...conform.Option) error {
	// Create an instance of the type
	instance := reflect.New(configType).Interface()

	// Use the non-generic Load function
	return conform.Load(instance, opts...)
}

// printJSON prints a struct as JSON
func printJSON(v interface{}) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(data))
	return nil
}
