package conform

import (
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/alicanli1995/conform/convert"
	"github.com/alicanli1995/conform/loader"
	"github.com/alicanli1995/conform/parser"
	"github.com/alicanli1995/conform/source"
	"github.com/alicanli1995/conform/validate"
)

// FieldError represents an error for a specific field
type FieldError struct {
	Field      string
	FieldPath  string
	Value      interface{}
	Message    string
	Key        string // env var or file key
	Location   string // where the error occurred
	Suggestion string
}

// ErrorList collects multiple field errors
type ErrorList struct {
	Errors []FieldError
}

func (e *ErrorList) Error() string {
	if len(e.Errors) == 0 {
		return "no errors"
	}

	var sb strings.Builder
	sb.WriteString("❌ Configuration validation failed:\n\n")

	for i, err := range e.Errors {
		sb.WriteString(fmt.Sprintf("%d. %s", i+1, err.FieldPath))
		if err.Key != "" {
			sb.WriteString(fmt.Sprintf(" (%s)", err.Key))
		}
		sb.WriteString(fmt.Sprintf(": %s\n", err.Message))

		if err.Value != nil {
			sb.WriteString(fmt.Sprintf("   Got: %v\n", err.Value))
		}

		if err.Location != "" {
			sb.WriteString(fmt.Sprintf("   Location: %s\n", err.Location))
		}

		if err.Suggestion != "" {
			sb.WriteString(fmt.Sprintf("   💡 Suggestion: %s\n", err.Suggestion))
		}

		if i < len(e.Errors)-1 {
			sb.WriteString("\n")
		}
	}

	return sb.String()
}

// Config holds the configuration for loading
type Config struct {
	Source      source.Source
	FileLoader  *loader.FileLoader
	Converter   *convert.Converter
	Validator   *validate.Validator
	FileSources []source.Source
	Sources     []source.Source
	Environment string            // Environment name (dev, staging, production, etc.)
	Variables   map[string]string // Custom variables for substitution
}

// Option is a configuration option
type Option func(*Config)

// FromEnv creates an option that loads from environment variables
func FromEnv() Option {
	return func(c *Config) {
		c.Sources = append(c.Sources, source.NewEnvSource())
	}
}

// FromFile creates an option that loads from a file
// Supports environment-specific files: config.${ENV}.yaml
func FromFile(filePath string) Option {
	return func(c *Config) {
		if filePath != "" {
			// Substitute environment variables in file path
			resolvedPath := substituteVariables(filePath, c.Environment, c.Variables)

			fileLoader := loader.NewFileLoader(resolvedPath)
			c.FileLoader = fileLoader
			if err := fileLoader.Load(); err != nil {
				// Log error but don't fail
				_ = err
			}
			fileSource := &fileSource{loader: fileLoader, path: resolvedPath}
			c.Sources = append(c.Sources, fileSource)
		}
	}
}

// FromConsul creates an option that loads from Consul (placeholder - would need consul client)
func FromConsul(path string) Option {
	return func(c *Config) {
		// Placeholder - would integrate with Consul client
		// For now, we'll create a mock source
		consulSource := &consulSource{path: path}
		c.Sources = append(c.Sources, consulSource)
	}
}

// WithSource sets a custom source
func WithSource(s source.Source) Option {
	return func(c *Config) {
		c.Sources = append(c.Sources, s)
	}
}

// WithFileLoader sets a file loader
func WithFileLoader(fileLoader *loader.FileLoader) Option {
	return func(c *Config) {
		c.FileLoader = fileLoader
		if err := fileLoader.Load(); err != nil {
			_ = err
		}
		fileSource := &fileSource{loader: fileLoader}
		c.Sources = append(c.Sources, fileSource)
	}
}

// WithConverter sets a custom converter
func WithConverter(conv *convert.Converter) Option {
	return func(c *Config) {
		c.Converter = conv
	}
}

// WithValidator sets a custom validator
func WithValidator(v *validate.Validator) Option {
	return func(c *Config) {
		c.Validator = v
	}
}

// WithEnvironment sets the environment name for environment-specific configs
// Example: WithEnvironment("production") will load config.production.yaml
func WithEnvironment(env string) Option {
	return func(c *Config) {
		c.Environment = env
		// Also set ENV variable for substitution
		if c.Variables == nil {
			c.Variables = make(map[string]string)
		}
		c.Variables["ENV"] = env
	}
}

// WithVariables sets custom variables for substitution in config values
// Example: WithVariables(map[string]string{"APP_NAME": "MyApp"})
func WithVariables(vars map[string]string) Option {
	return func(c *Config) {
		if c.Variables == nil {
			c.Variables = make(map[string]string)
		}
		for k, v := range vars {
			c.Variables[k] = v
		}
	}
}

// Load loads configuration into the target struct (non-generic version for backward compatibility)
func Load(target interface{}, opts ...Option) error {
	targetType := reflect.TypeOf(target)
	if targetType.Kind() != reflect.Ptr {
		return fmt.Errorf("target must be a pointer to struct")
	}

	elemType := targetType.Elem()
	if elemType.Kind() != reflect.Struct {
		return fmt.Errorf("target must be a pointer to struct")
	}

	// Create new instance
	newValue := reflect.New(elemType)

	// Load into new instance
	err := loadInto(newValue.Interface(), opts...)
	if err != nil {
		return err
	}

	// Copy to target
	reflect.ValueOf(target).Elem().Set(newValue.Elem())
	return nil
}

// LoadGeneric loads configuration using generics - returns the config struct directly
func LoadGeneric[T any](opts ...Option) (*T, error) {
	var zero T
	targetType := reflect.TypeOf(zero)
	if targetType.Kind() != reflect.Struct {
		return nil, fmt.Errorf("T must be a struct type")
	}

	// Create new instance
	result := new(T)

	err := loadInto(result, opts...)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// loadInto is the internal implementation
func loadInto(target interface{}, opts ...Option) error {
	// Default configuration
	cfg := &Config{
		Converter: defaultConverter,
		Validator: defaultValidator,
		Sources:   []source.Source{source.NewEnvSource()},
		Variables: make(map[string]string),
	}

	// Apply options
	for _, opt := range opts {
		opt(cfg)
	}

	// Build multi-source (priority: first source wins)
	multiSource := source.NewMultiSource(cfg.Sources...)

	// Parse struct
	targetType := reflect.TypeOf(target)
	if targetType.Kind() == reflect.Ptr {
		targetType = targetType.Elem()
	}

	info, err := parser.ParseStruct(targetType)
	if err != nil {
		return fmt.Errorf("failed to parse struct: %w", err)
	}

	// Get target value
	targetValue := reflect.ValueOf(target)
	if targetValue.Kind() == reflect.Ptr {
		targetValue = targetValue.Elem()
	}

	var errorList ErrorList

	// Process each field
	for _, fieldInfo := range info.Fields {
		fieldValue := targetValue.FieldByIndex(fieldInfo.Index)
		fieldPath := buildFieldPath(fieldInfo, info, targetType)

		// Get value from source
		var value string
		var found bool
		var sourceName string

		// Try env var first
		if fieldInfo.Tag.EnvVar != "" {
			value, found = multiSource.Get(fieldInfo.Tag.EnvVar)
			if found {
				sourceName = fmt.Sprintf("env var %s", fieldInfo.Tag.EnvVar)
			}
		}

		// Try file key if not found
		if !found && fieldInfo.Tag.FileKey != "" && cfg.FileLoader != nil {
			// Substitute variables in file key (e.g., database.${ENV}.host -> database.production.host)
			resolvedFileKey := substituteVariables(fieldInfo.Tag.FileKey, cfg.Environment, cfg.Variables)
			value, found = cfg.FileLoader.Get(resolvedFileKey)
			if found {
				sourceName = fmt.Sprintf("file key %s", resolvedFileKey)
			}
		}

		// Use default if not found
		if !found && fieldInfo.Tag.Default != "" {
			value = fieldInfo.Tag.Default
			found = true
			sourceName = "default value"
		}

		// Substitute variables in value (${VAR_NAME:-default})
		if found {
			value = substituteVariables(value, cfg.Environment, cfg.Variables)
		}

		// Check required
		if fieldInfo.Tag.Required && !found {
			var keyName string
			if fieldInfo.Tag.EnvVar != "" {
				keyName = fieldInfo.Tag.EnvVar
			} else if fieldInfo.Tag.FileKey != "" {
				keyName = fieldInfo.Tag.FileKey
			} else {
				keyName = fieldInfo.Name
			}

			suggestion := fmt.Sprintf("Set via: %s environment variable", keyName)
			if fieldInfo.Tag.FileKey != "" {
				suggestion = fmt.Sprintf("Set via: %s file key or %s environment variable", fieldInfo.Tag.FileKey, keyName)
			}

			errorList.Errors = append(errorList.Errors, FieldError{
				Field:      fieldInfo.Name,
				FieldPath:  fieldPath,
				Key:        keyName,
				Message:    "missing required field",
				Suggestion: suggestion,
			})
			continue
		}

		// Convert value
		if found {
			// Use separator from tag for slices, format for time and other structs
			var formatOrSeparator string
			if fieldInfo.Type.Kind() == reflect.Slice {
				formatOrSeparator = fieldInfo.Tag.Separator
			} else if fieldInfo.Type.Kind() == reflect.Struct {
				// Check if it's time.Time by comparing package path and name
				if fieldInfo.Type.PkgPath() == "time" && fieldInfo.Type.Name() == "Time" {
					formatOrSeparator = fieldInfo.Tag.Format
				} else {
					formatOrSeparator = fieldInfo.Tag.Format
				}
			} else {
				formatOrSeparator = fieldInfo.Tag.Format
			}
			converted, err := cfg.Converter.Convert(value, fieldInfo.Type, formatOrSeparator)
			if err != nil {
				suggestion := getConversionSuggestion(fieldInfo.Type, value)
				errorList.Errors = append(errorList.Errors, FieldError{
					Field:      fieldInfo.Name,
					FieldPath:  fieldPath,
					Value:      value,
					Key:        fieldInfo.Tag.EnvVar,
					Message:    fmt.Sprintf("conversion failed: %v", err),
					Location:   sourceName,
					Suggestion: suggestion,
				})
				continue
			}

			fieldValue.Set(reflect.ValueOf(converted))
		}

		// Validate
		if len(fieldInfo.Tag.Validators) > 0 {
			fieldVal := fieldValue.Interface()
			err := cfg.Validator.Validate(fieldVal, fieldInfo.Tag.Validators)
			if err != nil {
				suggestion := getValidationSuggestion(fieldInfo.Tag.Validators, fieldInfo.Type)
				errorList.Errors = append(errorList.Errors, FieldError{
					Field:      fieldInfo.Name,
					FieldPath:  fieldPath,
					Value:      fieldVal,
					Key:        fieldInfo.Tag.EnvVar,
					Message:    err.Error(),
					Location:   sourceName,
					Suggestion: suggestion,
				})
			}
		}
	}

	if len(errorList.Errors) > 0 {
		return &errorList
	}

	return nil
}

// buildFieldPath builds the full path to a field (e.g., "Database.URL")
func buildFieldPath(fieldInfo parser.FieldInfo, structInfo *parser.StructInfo, targetType reflect.Type) string {
	if len(fieldInfo.Index) == 1 {
		// Top-level field
		return fieldInfo.Name
	}

	// Build path from field index
	var path []string
	currentType := targetType

	for i := 0; i < len(fieldInfo.Index)-1; i++ {
		idx := fieldInfo.Index[i]
		if idx < currentType.NumField() {
			field := currentType.Field(idx)
			path = append(path, field.Name)

			// Navigate to nested struct
			if field.Type.Kind() == reflect.Ptr {
				currentType = field.Type.Elem()
			} else {
				currentType = field.Type
			}

			if currentType.Kind() != reflect.Struct {
				break
			}
		}
	}

	// Add the final field name
	path = append(path, fieldInfo.Name)

	return strings.Join(path, ".")
}

// getConversionSuggestion provides helpful suggestions for conversion errors
func getConversionSuggestion(targetType reflect.Type, value string) string {
	switch targetType.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return fmt.Sprintf("Provide a valid integer (got: %q)", value)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return fmt.Sprintf("Provide a valid positive integer (got: %q)", value)
	case reflect.Float32, reflect.Float64:
		return fmt.Sprintf("Provide a valid number (got: %q)", value)
	case reflect.Bool:
		return fmt.Sprintf("Provide 'true', 'false', '1', '0', 'yes', 'no' (got: %q)", value)
	}
	return fmt.Sprintf("Check the format of the value: %q", value)
}

// getValidationSuggestion provides helpful suggestions for validation errors
func getValidationSuggestion(validators []parser.Validator, fieldType reflect.Type) string {
	var suggestions []string

	for _, v := range validators {
		switch v.Name {
		case "gte":
			if len(v.Params) > 0 {
				suggestions = append(suggestions, fmt.Sprintf("Use a value >= %s", v.Params[0]))
			}
		case "lte":
			if len(v.Params) > 0 {
				suggestions = append(suggestions, fmt.Sprintf("Use a value <= %s", v.Params[0]))
			}
		case "min":
			if len(v.Params) > 0 {
				suggestions = append(suggestions, fmt.Sprintf("Use a value >= %s", v.Params[0]))
			}
		case "max":
			if len(v.Params) > 0 {
				suggestions = append(suggestions, fmt.Sprintf("Use a value <= %s", v.Params[0]))
			}
		case "email":
			suggestions = append(suggestions, "Format should be: user@example.com")
		case "url":
			suggestions = append(suggestions, "Format should be: https://example.com")
		case "hostname":
			suggestions = append(suggestions, "Format should be: example.com")
		case "ip":
			suggestions = append(suggestions, "Format should be: 192.168.1.1 or ::1")
		}
	}

	if len(suggestions) > 0 {
		return strings.Join(suggestions, ", ")
	}
	return ""
}

// substituteVariables substitutes variables in a string
// Supports:
//   - ${VAR_NAME} - simple substitution
//   - ${VAR_NAME:-default} - substitution with default value
//   - ${ENV} - environment variable substitution
func substituteVariables(s string, env string, vars map[string]string) string {
	if s == "" {
		return s
	}

	// Build variable map from environment variables and custom variables
	allVars := make(map[string]string)

	// Add environment variables
	for _, envVar := range os.Environ() {
		parts := strings.SplitN(envVar, "=", 2)
		if len(parts) == 2 {
			allVars[parts[0]] = parts[1]
		}
	}

	// Add custom variables (override env vars)
	if vars != nil {
		for k, v := range vars {
			allVars[k] = v
		}
	}

	// Add ENV variable if set
	if env != "" {
		allVars["ENV"] = env
	}

	// Replace ${VAR_NAME} and ${VAR_NAME:-default}
	var result strings.Builder
	i := 0
	for i < len(s) {
		if s[i] == '$' && i+1 < len(s) && s[i+1] == '{' {
			// Found ${ - find closing }
			start := i + 2
			end := start
			for end < len(s) && s[end] != '}' {
				end++
			}

			if end < len(s) {
				varExpr := s[start:end]

				// Check for default value syntax: VAR_NAME:-default
				if idx := strings.Index(varExpr, ":-"); idx != -1 {
					varName := strings.TrimSpace(varExpr[:idx])
					defaultVal := strings.TrimSpace(varExpr[idx+2:])

					if val, ok := allVars[varName]; ok && val != "" {
						result.WriteString(val)
					} else {
						result.WriteString(defaultVal)
					}
				} else {
					// Simple substitution
					varName := strings.TrimSpace(varExpr)
					if val, ok := allVars[varName]; ok {
						result.WriteString(val)
					} else {
						// Variable not found, keep original
						result.WriteString("${" + varExpr + "}")
					}
				}

				i = end + 1
			} else {
				// No closing brace, keep as is
				result.WriteByte(s[i])
				i++
			}
		} else {
			result.WriteByte(s[i])
			i++
		}
	}

	return result.String()
}

// fileSource wraps FileLoader to implement Source interface
type fileSource struct {
	loader *loader.FileLoader
	path   string
}

func (f *fileSource) Get(key string) (string, bool) {
	return f.loader.Get(key)
}

// consulSource is a placeholder for Consul integration
type consulSource struct {
	path string
}

func (c *consulSource) Get(key string) (string, bool) {
	// Placeholder - would integrate with Consul client
	return "", false
}

// Default converter and validator instances
var (
	defaultConverter = convert.New()
	defaultValidator = validate.New()
)

// RegisterValidator registers a global validator
func RegisterValidator(name string, fn func(value interface{}, params []string) error) {
	defaultValidator.RegisterValidator(name, fn)
}

// RegisterConverter registers a global converter
func RegisterConverter(t reflect.Type, fn convert.ConvertFunc) {
	defaultConverter.RegisterConverter(t, fn)
}

// WithFile sets a config file path (backward compatibility alias for FromFile)
func WithFile(filePath string) Option {
	return FromFile(filePath)
}
