package parser

import (
	"fmt"
	"reflect"
	"strings"
)

type TagSpec struct {
	EnvVar     string
	FileKey    string
	Default    string
	Required   bool
	Prefix     string
	Separator  string
	Format     string
	Validators []Validator
}

type Validator struct {
	Name   string
	Params []string
}

func ParseTag(tag string) (*TagSpec, error) {
	spec := &TagSpec{
		Separator: ",",
	}

	if tag == "" {
		return spec, nil
	}

	parts := splitTag(tag)

	for _, part := range parts {
		key, value := splitKeyValue(part)

		switch key {
		case "env":
			spec.EnvVar = value

		case "file":
			spec.FileKey = value

		case "default":
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
			validators := parseValidators(value)
			spec.Validators = append(spec.Validators, validators...)

		default:
			return nil, fmt.Errorf("unknown tag key: %s", key)
		}
	}

	return spec, nil
}

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
			partStr := current.String()
			if strings.HasPrefix(partStr, "validate=") {
				current.WriteByte(ch)
				i++
			} else {
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

func splitKeyValue(s string) (string, string) {
	if idx := strings.Index(s, "="); idx != -1 {
		return strings.TrimSpace(s[:idx]), strings.TrimSpace(s[idx+1:])
	}

	if idx := strings.Index(s, ":"); idx != -1 {
		return strings.TrimSpace(s[:idx]), strings.TrimSpace(s[idx+1:])
	}

	return s, ""
}

func parseValidators(s string) []Validator {
	parts := strings.Split(s, ",")
	validators := make([]Validator, 0, len(parts))

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		colonIdx := strings.Index(part, ":")
		if colonIdx == -1 {
			validators = append(validators, Validator{
				Name: part,
			})
		} else {
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

type StructInfo struct {
	Fields []FieldInfo
}

type FieldInfo struct {
	Name       string
	Type       reflect.Type
	Tag        *TagSpec
	Index      []int
	IsEmbedded bool
	FieldName  string
}

func ParseStruct(t reflect.Type) (*StructInfo, error) {
	if t.Kind() != reflect.Struct {
		return nil, fmt.Errorf("expected struct, got %s", t.Kind())
	}

	info := &StructInfo{
		Fields: make([]FieldInfo, 0),
	}

	parseStructFields(t, nil, "", info)

	return info, nil
}

func parseStructFields(t reflect.Type, index []int, prefix string, info *StructInfo) {
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		if !field.IsExported() {
			continue
		}

		fieldIndex := append(index, i)

		tagStr := field.Tag.Get("conform")
		tag, err := ParseTag(tagStr)
		if err != nil {
			continue
		}

		if field.Anonymous && field.Type.Kind() == reflect.Struct {
			embedPrefix := prefix
			if tag.Prefix != "" {
				embedPrefix = prefix + tag.Prefix
			}
			parseStructFields(field.Type, fieldIndex, embedPrefix, info)
			continue
		}

		if field.Type.Kind() == reflect.Struct && tag.Prefix != "" {
			nestedPrefix := prefix + tag.Prefix
			parseStructFields(field.Type, fieldIndex, nestedPrefix, info)
			continue
		}

		if field.Type.Kind() == reflect.Struct && tag.FileKey == "" &&
			!(field.Type.PkgPath() == "time" && field.Type.Name() == "Time") {
			parseStructFields(field.Type, fieldIndex, prefix, info)
			continue
		}

		if prefix != "" && tag.EnvVar != "" {
			tag.EnvVar = prefix + tag.EnvVar
		}

		info.Fields = append(info.Fields, FieldInfo{
			Name:      field.Name,
			Type:      field.Type,
			Tag:       tag,
			Index:     fieldIndex,
			FieldName: field.Name,
		})
	}
}
