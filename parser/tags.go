package parser

import (
	"fmt"
	"reflect"
	"strings"
)

// TagSpec represents parsed tag information
type TagSpec struct {
	// Source configuration
	EnvVar   string // Environment variable name
	FileKey  string // Key in config file (for nested configs)
	Default  string // Default value if not found
	Required bool   // Is this field required?
	Prefix   string // Prefix for nested structs (e.g. "DB_")

	// Type conversion
	Separator string // For slices: "1,2,3" with separator ","
	Format    string // For time.Time: "2006-01-02"

	// Validation rules
	Validators []Validator
}

// Validator represents a validation rule
type Validator struct {
	Name   string   // "min", "max", "email", "url"
	Params []string // Parameters for validator
}

// ParseTag parses the conform struct tag
// Format: conform:"env=PORT,default=8080,validate=gte:1024"
func ParseTag(tag string) (*TagSpec, error) {
	spec := &TagSpec{
		Separator: ",", // Default separator for slices
	}

	if tag == "" {
		return spec, nil
	}

	// Split by comma, but respect escaped commas and quotes
	parts := splitTag(tag)

	for _, part := range parts {
		key, value := splitKeyValue(part)

		switch key {
		case "env":
			spec.EnvVar = value

		case "file":
			spec.FileKey = value

		case "default":
			// Remove surrounding quotes if present
			if len(value) >= 2 && value[0] == '"' && value[len(value)-1] == '"' {
				spec.Default = value[1 : len(value)-1]
			} else {
				spec.Default = value
			}

		case "required":
			spec.Required = true

		case "prefix":
			spec.Prefix = value

		case "separator":
			spec.Separator = value

		case "format":
			spec.Format = value

		case "validate":
			// Parse validation rules: "gte:1024,lte:65535,numeric"
			validators := parseValidators(value)
			spec.Validators = append(spec.Validators, validators...)

		default:
			return nil, fmt.Errorf("unknown tag key: %s", key)
		}
	}

	return spec, nil
}

// splitTag splits tag by comma, respecting quotes, escapes, and validate values
func splitTag(tag string) []string {
	var parts []string
	tag = strings.TrimSpace(tag)
	if tag == "" {
		return parts
	}

	var current strings.Builder
	inQuote := false
	escaped := false
	i := 0

	for i < len(tag) {
		ch := tag[i]

		switch {
		case escaped:
			current.WriteByte(ch)
			escaped = false
			i++

		case ch == '\\':
			escaped = true
			current.WriteByte(ch)
			i++

		case ch == '"':
			inQuote = !inQuote
			current.WriteByte(ch)
			i++

		case ch == ',' && !inQuote:
			// Check if we're inside a validate= value
			// Look back to see if we have "validate=" before us
			partStr := current.String()
			if strings.HasPrefix(partStr, "validate=") {
				// We're in validate, don't split - just continue
				current.WriteByte(ch)
				i++
			} else {
				// Normal split
				if current.Len() > 0 {
					parts = append(parts, strings.TrimSpace(partStr))
					current.Reset()
				}
				i++
			}

		default:
			current.WriteByte(ch)
			i++
		}
	}

	if current.Len() > 0 {
		parts = append(parts, strings.TrimSpace(current.String()))
	}

	return parts
}

// splitKeyValue splits "key=value" or "key:value"
func splitKeyValue(s string) (string, string) {
	// Try '=' first
	if idx := strings.Index(s, "="); idx != -1 {
		return strings.TrimSpace(s[:idx]), strings.TrimSpace(s[idx+1:])
	}

	// Try ':' for validators
	if idx := strings.Index(s, ":"); idx != -1 {
		return strings.TrimSpace(s[:idx]), strings.TrimSpace(s[idx+1:])
	}

	// No value, just key (like "required")
	return s, ""
}

// parseValidators parses validation rules
// Examples: "min:3,max:20,email"
func parseValidators(s string) []Validator {
	parts := strings.Split(s, ",")
	validators := make([]Validator, 0, len(parts))

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		// Split by colon for parameters
		colonIdx := strings.Index(part, ":")
		if colonIdx == -1 {
			// No parameters
			validators = append(validators, Validator{
				Name: part,
			})
		} else {
			// Has parameters: "min:3" or "between:1:10"
			name := part[:colonIdx]
			params := strings.Split(part[colonIdx+1:], ":")
			validators = append(validators, Validator{
				Name:   name,
				Params: params,
			})
		}
	}

	return validators
}

// StructInfo holds metadata about a struct
type StructInfo struct {
	Fields []FieldInfo
}

// FieldInfo holds information about a struct field
type FieldInfo struct {
	Name       string
	Type       reflect.Type
	Tag        *TagSpec
	Index      []int // Field index path for embedded structs
	IsEmbedded bool
	FieldName  string // Original struct field name
}

// ParseStruct extracts all fields with conform tags
func ParseStruct(t reflect.Type) (*StructInfo, error) {
	if t.Kind() != reflect.Struct {
		return nil, fmt.Errorf("expected struct, got %s", t.Kind())
	}

	info := &StructInfo{
		Fields: make([]FieldInfo, 0),
	}

	// Recursively parse fields
	parseStructFields(t, nil, "", info)

	return info, nil
}

func parseStructFields(t reflect.Type, index []int, prefix string, info *StructInfo) {
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		// Get field index path
		fieldIndex := append(index, i)

		// Parse tag
		tagStr := field.Tag.Get("conform")
		tag, err := ParseTag(tagStr)
		if err != nil {
			// Log warning but continue
			continue
		}

		// Handle embedded structs
		if field.Anonymous && field.Type.Kind() == reflect.Struct {
			// Recursively parse embedded struct
			embedPrefix := prefix
			if tag.Prefix != "" {
				embedPrefix = prefix + tag.Prefix
			}
			parseStructFields(field.Type, fieldIndex, embedPrefix, info)
			continue
		}

		// Handle nested structs with prefix (for env vars)
		if field.Type.Kind() == reflect.Struct && tag.Prefix != "" {
			nestedPrefix := prefix + tag.Prefix
			parseStructFields(field.Type, fieldIndex, nestedPrefix, info)
			continue
		}

		// Handle nested structs with file keys (recursively parse without prefix)
		if field.Type.Kind() == reflect.Struct && tag.FileKey == "" &&
			!(field.Type.PkgPath() == "time" && field.Type.Name() == "Time") {
			parseStructFields(field.Type, fieldIndex, prefix, info)
			continue
		}

		// Apply prefix to env var
		if prefix != "" && tag.EnvVar != "" {
			tag.EnvVar = prefix + tag.EnvVar
		}

		// Add field info
		info.Fields = append(info.Fields, FieldInfo{
			Name:      field.Name,
			Type:      field.Type,
			Tag:       tag,
			Index:     fieldIndex,
			FieldName: field.Name,
		})
	}
}
