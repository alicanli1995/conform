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

func (f *FileLoader) Load() error {
	if f.filePath == "" {
		return nil
	}

	data, err := os.ReadFile(f.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
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

func convertToMap(data interface{}) map[string]interface{} {
	if data == nil {
		return make(map[string]interface{})
	}

	if m, ok := data.(map[string]interface{}); ok {
		return m
	}

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

		if i == len(parts)-1 {
			return fmt.Sprintf("%v", value), true
		}

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
