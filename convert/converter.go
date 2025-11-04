package convert

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// Converter converts string values to typed values
type Converter struct {
	customConverters map[reflect.Type]ConvertFunc
}

// ConvertFunc is a function that converts a string to a typed value
type ConvertFunc func(string) (interface{}, error)

// New creates a new Converter instance
func New() *Converter {
	return &Converter{
		customConverters: make(map[reflect.Type]ConvertFunc),
	}
}

// RegisterConverter registers a custom converter for a type
func (c *Converter) RegisterConverter(t reflect.Type, fn ConvertFunc) {
	c.customConverters[t] = fn
}

func (c *Converter) Convert(value string, targetType reflect.Type, format string) (interface{}, error) {
	if fn, ok := c.customConverters[targetType]; ok {
		return fn(value)
	}

	if targetType.Kind() == reflect.Ptr {
		elemType := targetType.Elem()
		converted, err := c.Convert(value, elemType, format)
		if err != nil {
			return nil, err
		}

		ptr := reflect.New(elemType)
		ptr.Elem().Set(reflect.ValueOf(converted))
		return ptr.Interface(), nil
	}

	switch targetType.Kind() {
	case reflect.String:
		return value, nil

	case reflect.Bool:
		return parseBool(value)

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if targetType == reflect.TypeOf(time.Duration(0)) {
			return time.ParseDuration(value)
		}
		return parseInt(value, targetType)

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return parseUint(value, targetType)

	case reflect.Float32, reflect.Float64:
		return parseFloat(value, targetType)

	case reflect.Slice:
		return c.convertSlice(value, targetType, format)

	case reflect.Map:
		return c.convertMap(value, targetType)

	case reflect.Struct:
		return c.convertStruct(value, targetType, format)

	default:
		return nil, fmt.Errorf("unsupported type: %s", targetType)
	}
}

func parseBool(value string) (bool, error) {
	value = strings.ToLower(strings.TrimSpace(value))

	switch value {
	case "true", "t", "yes", "y", "1", "on":
		return true, nil
	case "false", "f", "no", "n", "0", "off", "":
		return false, nil
	default:
		return false, fmt.Errorf("invalid bool value: %s", value)
	}
}

func parseInt(value string, targetType reflect.Type) (interface{}, error) {
	bitSize := targetType.Bits()
	i, err := strconv.ParseInt(value, 10, bitSize)
	if err != nil {
		return nil, fmt.Errorf("invalid integer: %s", value)
	}

	// Convert to exact type
	switch targetType.Kind() {
	case reflect.Int:
		return int(i), nil
	case reflect.Int8:
		return int8(i), nil
	case reflect.Int16:
		return int16(i), nil
	case reflect.Int32:
		return int32(i), nil
	case reflect.Int64:
		return int64(i), nil
	}

	return nil, fmt.Errorf("unexpected int type")
}

func parseUint(value string, targetType reflect.Type) (interface{}, error) {
	bitSize := targetType.Bits()
	u, err := strconv.ParseUint(value, 10, bitSize)
	if err != nil {
		return nil, fmt.Errorf("invalid unsigned integer: %s", value)
	}

	switch targetType.Kind() {
	case reflect.Uint:
		return uint(u), nil
	case reflect.Uint8:
		return uint8(u), nil
	case reflect.Uint16:
		return uint16(u), nil
	case reflect.Uint32:
		return uint32(u), nil
	case reflect.Uint64:
		return uint64(u), nil
	}

	return nil, fmt.Errorf("unexpected uint type")
}

func parseFloat(value string, targetType reflect.Type) (interface{}, error) {
	bitSize := targetType.Bits()
	f, err := strconv.ParseFloat(value, bitSize)
	if err != nil {
		return nil, fmt.Errorf("invalid float: %s", value)
	}

	if targetType.Kind() == reflect.Float32 {
		return float32(f), nil
	}
	return f, nil
}

func (c *Converter) convertSlice(value string, targetType reflect.Type, separator string) (interface{}, error) {
	if value == "" {
		return reflect.MakeSlice(targetType, 0, 0).Interface(), nil
	}

	if separator == "" {
		separator = ","
	}

	parts := strings.Split(value, separator)
	elemType := targetType.Elem()

	slice := reflect.MakeSlice(targetType, len(parts), len(parts))

	for i, part := range parts {
		part = strings.TrimSpace(part)

		elem, err := c.Convert(part, elemType, "")
		if err != nil {
			return nil, fmt.Errorf("slice element %d: %w", i, err)
		}

		slice.Index(i).Set(reflect.ValueOf(elem))
	}

	return slice.Interface(), nil
}

func (c *Converter) convertMap(value string, targetType reflect.Type) (interface{}, error) {
	if value == "" {
		return reflect.MakeMap(targetType).Interface(), nil
	}

	m := reflect.MakeMap(targetType)
	keyType := targetType.Key()
	valueType := targetType.Elem()

	pairs := strings.Split(value, ",")
	for _, pair := range pairs {
		parts := strings.SplitN(pair, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid map pair: %s (expected format: key=value)", pair)
		}

		keyStr := strings.TrimSpace(parts[0])
		key, err := c.Convert(keyStr, keyType, "")
		if err != nil {
			return nil, fmt.Errorf("map key conversion failed for %q: %w", keyStr, err)
		}

		valStr := strings.TrimSpace(parts[1])
		val, err := c.Convert(valStr, valueType, "")
		if err != nil {
			return nil, fmt.Errorf("map value conversion failed for %q: %w", valStr, err)
		}

		m.SetMapIndex(reflect.ValueOf(key), reflect.ValueOf(val))
	}

	return m.Interface(), nil
}

func (c *Converter) convertStruct(value string, targetType reflect.Type, format string) (interface{}, error) {
	if targetType.PkgPath() == "time" && targetType.Name() == "Time" {
		layout := time.RFC3339
		if format != "" {
			layout = format
		}
		t, err := time.Parse(layout, value)
		if err != nil {
			return nil, fmt.Errorf("failed to parse time %q with layout %q: %w", value, layout, err)
		}
		return t, nil
	}

	return nil, fmt.Errorf("cannot convert string to struct %s", targetType)
}
