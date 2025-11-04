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
	Key        string
	Location   string
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
	Environment string
	Variables   map[string]string
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
func FromFile(filePath string) Option {
	return func(c *Config) {
		if filePath != "" {
			resolvedPath := substituteVariables(filePath, c.Environment, c.Variables)

			fileLoader := loader.NewFileLoader(resolvedPath)
			c.FileLoader = fileLoader
			if err := fileLoader.Load(); err != nil {
				_ = err
			}
			fileSource := &fileSource{loader: fileLoader, path: resolvedPath}
			c.Sources = append(c.Sources, fileSource)
		}
	}
}

// FromConsul creates an option that loads from Consul
func FromConsul(path string) Option {
	return func(c *Config) {
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
func WithEnvironment(env string) Option {
	return func(c *Config) {
		c.Environment = env
		if c.Variables == nil {
			c.Variables = make(map[string]string)
		}
		c.Variables["ENV"] = env
	}
}

// WithVariables sets custom variables for substitution in config values
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

// Load loads configuration into the target struct
func Load(target interface{}, opts ...Option) error {
	targetType := reflect.TypeOf(target)
	if targetType.Kind() != reflect.Ptr {
		return fmt.Errorf("target must be a pointer to struct")
	}

	elemType := targetType.Elem()
	if elemType.Kind() != reflect.Struct {
		return fmt.Errorf("target must be a pointer to struct")
	}

	newValue := reflect.New(elemType)
	err := loadInto(newValue.Interface(), opts...)
	if err != nil {
		return err
	}

	reflect.ValueOf(target).Elem().Set(newValue.Elem())
	return nil
}

// LoadGeneric loads configuration using generics
func LoadGeneric[T any](opts ...Option) (*T, error) {
	var zero T
	targetType := reflect.TypeOf(zero)
	if targetType.Kind() != reflect.Struct {
		return nil, fmt.Errorf("t must be a struct type")
	}

	result := new(T)
	err := loadInto(result, opts...)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// loadInto is the internal implementation
func loadInto(target interface{}, opts ...Option) error {
	cfg := &Config{
		Converter: defaultConverter,
		Validator: defaultValidator,
		Sources:   []source.Source{source.NewEnvSource()},
		Variables: make(map[string]string),
	}

	for _, opt := range opts {
		opt(cfg)
	}

	multiSource := source.NewMultiSource(cfg.Sources...)

	targetType := reflect.TypeOf(target)
	if targetType.Kind() == reflect.Ptr {
		targetType = targetType.Elem()
	}

	info, err := parser.ParseStruct(targetType)
	if err != nil {
		return fmt.Errorf("failed to parse struct: %w", err)
	}

	targetValue := reflect.ValueOf(target)
	if targetValue.Kind() == reflect.Ptr {
		targetValue = targetValue.Elem()
	}

	var errorList ErrorList

	for _, fieldInfo := range info.Fields {
		fieldValue := targetValue.FieldByIndex(fieldInfo.Index)
		fieldPath := buildFieldPath(fieldInfo, info, targetType)

		var value string
		var found bool
		var sourceName string

		if fieldInfo.Tag.EnvVar != "" {
			value, found = multiSource.Get(fieldInfo.Tag.EnvVar)
			if found {
				sourceName = fmt.Sprintf("env var %s", fieldInfo.Tag.EnvVar)
			}
		}

		if !found && fieldInfo.Tag.FileKey != "" && cfg.FileLoader != nil {
			resolvedFileKey := substituteVariables(fieldInfo.Tag.FileKey, cfg.Environment, cfg.Variables)
			value, found = cfg.FileLoader.Get(resolvedFileKey)
			if found {
				sourceName = fmt.Sprintf("file key %s", resolvedFileKey)
			}
		}

		if !found && fieldInfo.Tag.Default != "" {
			value = fieldInfo.Tag.Default
			found = true
			sourceName = "default value"
		}

		if found {
			value = substituteVariables(value, cfg.Environment, cfg.Variables)
		}

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

		if found {
			var formatOrSeparator string
			if fieldInfo.Type.Kind() == reflect.Slice {
				formatOrSeparator = fieldInfo.Tag.Separator
			} else if fieldInfo.Type.Kind() == reflect.Struct {
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

func buildFieldPath(fieldInfo parser.FieldInfo, structInfo *parser.StructInfo, targetType reflect.Type) string {
	if len(fieldInfo.Index) == 1 {
		return fieldInfo.Name
	}

	var path []string
	currentType := targetType

	for i := 0; i < len(fieldInfo.Index)-1; i++ {
		idx := fieldInfo.Index[i]
		if idx < currentType.NumField() {
			field := currentType.Field(idx)
			path = append(path, field.Name)

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

func substituteVariables(s string, env string, vars map[string]string) string {
	if s == "" {
		return s
	}

	allVars := make(map[string]string)

	for _, envVar := range os.Environ() {
		parts := strings.SplitN(envVar, "=", 2)
		if len(parts) == 2 {
			allVars[parts[0]] = parts[1]
		}
	}

	for k, v := range vars {
		allVars[k] = v
	}

	if env != "" {
		allVars["ENV"] = env
	}

	var result strings.Builder
	i := 0
	for i < len(s) {
		if s[i] == '$' && i+1 < len(s) && s[i+1] == '{' {
			start := i + 2
			end := start
			for end < len(s) && s[end] != '}' {
				end++
			}

			if end < len(s) {
				varExpr := s[start:end]

				if idx := strings.Index(varExpr, ":-"); idx != -1 {
					varName := strings.TrimSpace(varExpr[:idx])
					defaultVal := strings.TrimSpace(varExpr[idx+2:])

					if val, ok := allVars[varName]; ok && val != "" {
						result.WriteString(val)
					} else {
						result.WriteString(defaultVal)
					}
				} else {
					varName := strings.TrimSpace(varExpr)
					if val, ok := allVars[varName]; ok {
						result.WriteString(val)
					} else {
						result.WriteString("${" + varExpr + "}")
					}
				}

				i = end + 1
			} else {
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

type consulSource struct {
	path string
}

func (c *consulSource) Get(key string) (string, bool) {
	return "", false
}

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

func WithFile(filePath string) Option {
	return FromFile(filePath)
}
