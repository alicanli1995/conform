package loader

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"gopkg.in/yaml.v3"
)

// FileLoader loads configuration from files
type FileLoader struct {
	filePath string
	data     map[string]interface{}
}

// NewFileLoader creates a new file loader
func NewFileLoader(filePath string) *FileLoader {
	return &FileLoader{
		filePath: filePath,
		data:     make(map[string]interface{}),
	}
}

// Load loads the configuration file
func (f *FileLoader) Load() error {
	if f.filePath == "" {
		return nil // No file specified
	}

	data, err := os.ReadFile(f.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // File doesn't exist, return empty map
		}
		return fmt.Errorf("failed to read config file: %w", err)
	}

	ext := filepath.Ext(f.filePath)
	switch ext {
	case ".yaml", ".yml":
		if err := yaml.Unmarshal(data, &f.data); err != nil {
			return fmt.Errorf("failed to parse YAML: %w", err)
		}
	case ".json":
		if err := json.Unmarshal(data, &f.data); err != nil {
			return fmt.Errorf("failed to parse JSON: %w", err)
		}
	case ".toml":
		var tomlData map[string]interface{}
		if err := toml.Unmarshal(data, &tomlData); err != nil {
			return fmt.Errorf("failed to parse TOML: %w", err)
		}
		f.data = tomlData
	default:
		return fmt.Errorf("unsupported file format: %s (supported: .yaml, .yml, .json, .toml)", ext)
	}

	return nil
}

// convertToMap converts TOML data to map[string]interface{}
func convertToMap(data interface{}) map[string]interface{} {
	if data == nil {
		return make(map[string]interface{})
	}

	if m, ok := data.(map[string]interface{}); ok {
		return m
	}

	// If it's a map with interface{} keys, convert it
	if m, ok := data.(map[interface{}]interface{}); ok {
		result := make(map[string]interface{})
		for k, v := range m {
			key := fmt.Sprintf("%v", k)
			result[key] = convertValue(v)
		}
		return result
	}

	return make(map[string]interface{})
}

// convertValue recursively converts TOML values
func convertValue(v interface{}) interface{} {
	switch val := v.(type) {
	case map[interface{}]interface{}:
		result := make(map[string]interface{})
		for k, v := range val {
			key := fmt.Sprintf("%v", k)
			result[key] = convertValue(v)
		}
		return result
	case []interface{}:
		result := make([]interface{}, len(val))
		for i, item := range val {
			result[i] = convertValue(item)
		}
		return result
	default:
		return v
	}
}

// Get retrieves a value from the loaded configuration using dot notation
// Example: "database.host" -> data["database"]["host"]
func (f *FileLoader) Get(key string) (string, bool) {
	if f.data == nil {
		return "", false
	}

	parts := splitKey(key)
	current := f.data

	for i, part := range parts {
		value, ok := current[part]
		if !ok {
			return "", false
		}

		// If this is the last part, convert to string
		if i == len(parts)-1 {
			return fmt.Sprintf("%v", value), true
		}

		// Otherwise, continue traversing
		nested, ok := value.(map[string]interface{})
		if !ok {
			return "", false
		}
		current = nested
	}

	return "", false
}

// splitKey splits a dot-notation key into parts
func splitKey(key string) []string {
	if key == "" {
		return []string{}
	}
	return strings.Split(key, ".")
}
