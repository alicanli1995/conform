package validate

import (
	"fmt"
	"net"
	"net/mail"
	"net/url"
	"reflect"
	"regexp"
	"strconv"
	"unicode"

	"github.com/alicanli1995/conform/parser"
)

type Validator struct {
	customValidators map[string]ValidateFunc
}

type ValidateFunc func(value interface{}, params []string) error

func New() *Validator {
	v := &Validator{
		customValidators: make(map[string]ValidateFunc),
	}

	v.registerBuiltins()

	return v
}

func (v *Validator) RegisterValidator(name string, fn ValidateFunc) {
	v.customValidators[name] = fn
}

func (v *Validator) Validate(value interface{}, rules []parser.Validator) error {
	for _, rule := range rules {
		if err := v.validateRule(value, rule); err != nil {
			return fmt.Errorf("validation '%s' failed: %w", rule.Name, err)
		}
	}
	return nil
}

func (v *Validator) validateRule(value interface{}, rule parser.Validator) error {
	if fn, ok := v.customValidators[rule.Name]; ok {
		return fn(value, rule.Params)
	}
	switch rule.Name {
	case "required":
		return v.validateRequired(value)
	case "min":
		return v.validateMin(value, rule.Params)
	case "max":
		return v.validateMax(value, rule.Params)
	case "gte":
		return v.validateGte(value, rule.Params)
	case "lte":
		return v.validateLte(value, rule.Params)
	case "eq":
		return v.validateEq(value, rule.Params)
	case "ne":
		return v.validateNe(value, rule.Params)
	case "email":
		return v.validateEmail(value)
	case "url":
		return v.validateURL(value, rule.Params)
	case "ip":
		return v.validateIP(value)
	case "alphanum":
		return v.validateAlphaNum(value)
	case "alpha":
		return v.validateAlpha(value)
	case "numeric":
		return v.validateNumeric(value)
	case "regex":
		return v.validateRegex(value, rule.Params)
	case "oneof":
		return v.validateOneOf(value, rule.Params)
	case "len":
		return v.validateLen(value, rule.Params)
	case "hostname":
		return v.validateHostname(value)
	case "port":
		return v.validatePort(value)
	default:
		return fmt.Errorf("unknown validator: %s", rule.Name)
	}
}

func (v *Validator) registerBuiltins() {
	// Additional custom validators
	v.RegisterValidator("has_upper", func(value interface{}, params []string) error {
		str, ok := value.(string)
		if !ok {
			return fmt.Errorf("expected string")
		}

		for _, ch := range str {
			if unicode.IsUpper(ch) {
				return nil
			}
		}
		return fmt.Errorf("must contain uppercase letter")
	})

	v.RegisterValidator("has_lower", func(value interface{}, params []string) error {
		str, ok := value.(string)
		if !ok {
			return fmt.Errorf("expected string")
		}

		for _, ch := range str {
			if unicode.IsLower(ch) {
				return nil
			}
		}
		return fmt.Errorf("must contain lowercase letter")
	})

	v.RegisterValidator("has_digit", func(value interface{}, params []string) error {
		str, ok := value.(string)
		if !ok {
			return fmt.Errorf("expected string")
		}

		for _, ch := range str {
			if unicode.IsDigit(ch) {
				return nil
			}
		}
		return fmt.Errorf("must contain digit")
	})

	v.RegisterValidator("has_special", func(value interface{}, params []string) error {
		str, ok := value.(string)
		if !ok {
			return fmt.Errorf("expected string")
		}

		for _, ch := range str {
			if !unicode.IsLetter(ch) && !unicode.IsDigit(ch) && !unicode.IsSpace(ch) {
				return nil
			}
		}
		return fmt.Errorf("must contain special character")
	})
}

func (v *Validator) validateRequired(value interface{}) error {
	if value == nil {
		return fmt.Errorf("required field is nil")
	}

	val := reflect.ValueOf(value)

	switch val.Kind() {
	case reflect.String:
		if val.String() == "" {
			return fmt.Errorf("required field is empty")
		}
	case reflect.Slice, reflect.Map, reflect.Array:
		if val.Len() == 0 {
			return fmt.Errorf("required field is empty")
		}
	case reflect.Ptr:
		if val.IsNil() {
			return fmt.Errorf("required field is nil")
		}
	}

	return nil
}

func (v *Validator) validateMin(value interface{}, params []string) error {
	if len(params) == 0 {
		return fmt.Errorf("min requires parameter")
	}

	min, err := strconv.ParseFloat(params[0], 64)
	if err != nil {
		return fmt.Errorf("invalid min parameter: %s", params[0])
	}

	val := reflect.ValueOf(value)

	switch val.Kind() {
	case reflect.String:
		if float64(len(val.String())) < min {
			return fmt.Errorf("length %d is less than minimum %v", len(val.String()), min)
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if float64(val.Int()) < min {
			return fmt.Errorf("value %d is less than minimum %v", val.Int(), min)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if float64(val.Uint()) < min {
			return fmt.Errorf("value %d is less than minimum %v", val.Uint(), min)
		}
	case reflect.Float32, reflect.Float64:
		if val.Float() < min {
			return fmt.Errorf("value %f is less than minimum %v", val.Float(), min)
		}
	case reflect.Slice, reflect.Array:
		if float64(val.Len()) < min {
			return fmt.Errorf("length %d is less than minimum %v", val.Len(), min)
		}
	}

	return nil
}

func (v *Validator) validateMax(value interface{}, params []string) error {
	if len(params) == 0 {
		return fmt.Errorf("max requires parameter")
	}

	max, err := strconv.ParseFloat(params[0], 64)
	if err != nil {
		return fmt.Errorf("invalid max parameter: %s", params[0])
	}

	val := reflect.ValueOf(value)

	switch val.Kind() {
	case reflect.String:
		if float64(len(val.String())) > max {
			return fmt.Errorf("length %d exceeds maximum %v", len(val.String()), max)
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if float64(val.Int()) > max {
			return fmt.Errorf("value %d exceeds maximum %v", val.Int(), max)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if float64(val.Uint()) > max {
			return fmt.Errorf("value %d exceeds maximum %v", val.Uint(), max)
		}
	case reflect.Float32, reflect.Float64:
		if val.Float() > max {
			return fmt.Errorf("value %f exceeds maximum %v", val.Float(), max)
		}
	case reflect.Slice, reflect.Array:
		if float64(val.Len()) > max {
			return fmt.Errorf("length %d exceeds maximum %v", val.Len(), max)
		}
	}

	return nil
}

func (v *Validator) validateGte(value interface{}, params []string) error {
	// Greater than or equal
	if len(params) == 0 {
		return fmt.Errorf("gte requires parameter")
	}

	threshold, err := strconv.ParseFloat(params[0], 64)
	if err != nil {
		return fmt.Errorf("invalid gte parameter: %s", params[0])
	}

	val := reflect.ValueOf(value)

	var numValue float64
	switch val.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		numValue = float64(val.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		numValue = float64(val.Uint())
	case reflect.Float32, reflect.Float64:
		numValue = val.Float()
	default:
		return fmt.Errorf("gte only works with numeric types")
	}

	if numValue < threshold {
		return fmt.Errorf("value %v is less than %v", numValue, threshold)
	}

	return nil
}

func (v *Validator) validateLte(value interface{}, params []string) error {
	// Less than or equal
	if len(params) == 0 {
		return fmt.Errorf("lte requires parameter")
	}

	threshold, err := strconv.ParseFloat(params[0], 64)
	if err != nil {
		return fmt.Errorf("invalid lte parameter: %s", params[0])
	}

	val := reflect.ValueOf(value)

	var numValue float64
	switch val.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		numValue = float64(val.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		numValue = float64(val.Uint())
	case reflect.Float32, reflect.Float64:
		numValue = val.Float()
	default:
		return fmt.Errorf("lte only works with numeric types")
	}

	if numValue > threshold {
		return fmt.Errorf("value %v exceeds %v", numValue, threshold)
	}

	return nil
}

func (v *Validator) validateEq(value interface{}, params []string) error {
	if len(params) == 0 {
		return fmt.Errorf("eq requires parameter")
	}

	expected := params[0]
	actual := fmt.Sprintf("%v", value)

	if actual != expected {
		return fmt.Errorf("value '%s' does not equal '%s'", actual, expected)
	}

	return nil
}

func (v *Validator) validateNe(value interface{}, params []string) error {
	if len(params) == 0 {
		return fmt.Errorf("ne requires parameter")
	}

	forbidden := params[0]
	actual := fmt.Sprintf("%v", value)

	if actual == forbidden {
		return fmt.Errorf("value '%s' must not equal '%s'", actual, forbidden)
	}

	return nil
}

func (v *Validator) validateEmail(value interface{}) error {
	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("expected string")
	}

	_, err := mail.ParseAddress(str)
	if err != nil {
		return fmt.Errorf("invalid email address")
	}

	return nil
}

func (v *Validator) validateURL(value interface{}, params []string) error {
	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("expected string")
	}

	u, err := url.Parse(str)
	if err != nil {
		return fmt.Errorf("invalid URL")
	}

	if u.Scheme == "" || u.Host == "" {
		return fmt.Errorf("URL must have scheme and host")
	}

	// Check for HTTPS if specified
	if len(params) > 0 && params[0] == "https" {
		if u.Scheme != "https" {
			return fmt.Errorf("URL must use HTTPS")
		}
	}

	return nil
}

func (v *Validator) validateIP(value interface{}) error {
	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("expected string")
	}

	if net.ParseIP(str) == nil {
		return fmt.Errorf("invalid IP address")
	}

	return nil
}

func (v *Validator) validateAlphaNum(value interface{}) error {
	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("expected string")
	}

	for _, ch := range str {
		if !unicode.IsLetter(ch) && !unicode.IsDigit(ch) {
			return fmt.Errorf("must contain only letters and digits")
		}
	}

	return nil
}

func (v *Validator) validateAlpha(value interface{}) error {
	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("expected string")
	}

	for _, ch := range str {
		if !unicode.IsLetter(ch) {
			return fmt.Errorf("must contain only letters")
		}
	}

	return nil
}

func (v *Validator) validateNumeric(value interface{}) error {
	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("expected string")
	}

	for _, ch := range str {
		if !unicode.IsDigit(ch) {
			return fmt.Errorf("must contain only digits")
		}
	}

	return nil
}

func (v *Validator) validateRegex(value interface{}, params []string) error {
	if len(params) == 0 {
		return fmt.Errorf("regex requires pattern parameter")
	}

	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("expected string")
	}

	pattern := params[0]
	re, err := regexp.Compile(pattern)
	if err != nil {
		return fmt.Errorf("invalid regex pattern: %s", pattern)
	}

	if !re.MatchString(str) {
		return fmt.Errorf("value does not match pattern")
	}

	return nil
}

func (v *Validator) validateOneOf(value interface{}, params []string) error {
	if len(params) == 0 {
		return fmt.Errorf("oneof requires at least one parameter")
	}

	actual := fmt.Sprintf("%v", value)
	for _, param := range params {
		if actual == param {
			return nil
		}
	}

	return fmt.Errorf("value '%s' is not one of: %v", actual, params)
}

func (v *Validator) validateLen(value interface{}, params []string) error {
	if len(params) == 0 {
		return fmt.Errorf("len requires parameter")
	}

	expectedLen, err := strconv.Atoi(params[0])
	if err != nil {
		return fmt.Errorf("invalid len parameter: %s", params[0])
	}

	val := reflect.ValueOf(value)
	var actualLen int

	switch val.Kind() {
	case reflect.String, reflect.Slice, reflect.Array, reflect.Map:
		actualLen = val.Len()
	default:
		return fmt.Errorf("len only works with string, slice, array, or map")
	}

	if actualLen != expectedLen {
		return fmt.Errorf("length %d does not equal expected %d", actualLen, expectedLen)
	}

	return nil
}

func (v *Validator) validateHostname(value interface{}) error {
	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("expected string")
	}

	// Basic hostname validation
	if len(str) == 0 || len(str) > 253 {
		return fmt.Errorf("invalid hostname length")
	}

	// Check for valid characters
	for _, ch := range str {
		if !((ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') ||
			(ch >= '0' && ch <= '9') || ch == '-' || ch == '.') {
			return fmt.Errorf("invalid hostname character")
		}
	}

	return nil
}

func (v *Validator) validatePort(value interface{}) error {
	var port int

	switch val := value.(type) {
	case int:
		port = val
	case int8:
		port = int(val)
	case int16:
		port = int(val)
	case int32:
		port = int(val)
	case int64:
		port = int(val)
	case uint:
		port = int(val)
	case uint8:
		port = int(val)
	case uint16:
		port = int(val)
	case uint32:
		port = int(val)
	case uint64:
		port = int(val)
	case string:
		p, err := strconv.Atoi(val)
		if err != nil {
			return fmt.Errorf("invalid port format")
		}
		port = p
	default:
		return fmt.Errorf("port must be a number")
	}

	if port < 1 || port > 65535 {
		return fmt.Errorf("port must be between 1 and 65535")
	}

	return nil
}
