package source

import (
	"os"
)

// Source is an interface for configuration sources
type Source interface {
	Get(key string) (string, bool)
}

// EnvSource reads from environment variables
type EnvSource struct{}

// NewEnvSource creates a new environment variable source
func NewEnvSource() *EnvSource {
	return &EnvSource{}
}

// Get retrieves an environment variable
func (e *EnvSource) Get(key string) (string, bool) {
	value := os.Getenv(key)
	return value, value != ""
}

// MapSource is a map-based source (useful for testing or file-based configs)
type MapSource struct {
	data map[string]string
}

// NewMapSource creates a new map source
func NewMapSource(data map[string]string) *MapSource {
	return &MapSource{data: data}
}

// Get retrieves a value from the map
func (m *MapSource) Get(key string) (string, bool) {
	value, ok := m.data[key]
	return value, ok
}

// Set sets a value in the map
func (m *MapSource) Set(key, value string) {
	if m.data == nil {
		m.data = make(map[string]string)
	}
	m.data[key] = value
}

// MultiSource combines multiple sources, checking them in order
type MultiSource struct {
	sources []Source
}

// NewMultiSource creates a new multi-source
func NewMultiSource(sources ...Source) *MultiSource {
	return &MultiSource{sources: sources}
}

// Get retrieves a value from the first source that has it
func (m *MultiSource) Get(key string) (string, bool) {
	for _, source := range m.sources {
		if value, ok := source.Get(key); ok {
			return value, ok
		}
	}
	return "", false
}
