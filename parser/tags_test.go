package parser

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseSimpleTag(t *testing.T) {
	tag := "env=PORT,default=8080"
	spec, err := ParseTag(tag)

	assert.NoError(t, err)
	assert.Equal(t, "PORT", spec.EnvVar)
	assert.Equal(t, "8080", spec.Default)
}

func TestParseValidationTag(t *testing.T) {
	tag := "env=PORT,validate=gte:1024,lte:65535"
	spec, err := ParseTag(tag)

	assert.NoError(t, err)
	assert.Len(t, spec.Validators, 2)
	assert.Equal(t, "gte", spec.Validators[0].Name)
	assert.Equal(t, []string{"1024"}, spec.Validators[0].Params)
	assert.Equal(t, "lte", spec.Validators[1].Name)
	assert.Equal(t, []string{"65535"}, spec.Validators[1].Params)
}

func TestParseStructWithPrefix(t *testing.T) {
	type DBConfig struct {
		Host string `conform:"env=HOST"`
		Port int    `conform:"env=PORT"`
	}

	type Config struct {
		Database DBConfig `conform:"prefix=DB_"`
	}

	info, err := ParseStruct(reflect.TypeOf(Config{}))
	assert.NoError(t, err)

	// Should have 2 fields with DB_ prefix
	assert.Len(t, info.Fields, 2)
	assert.Equal(t, "DB_HOST", info.Fields[0].Tag.EnvVar)
	assert.Equal(t, "DB_PORT", info.Fields[1].Tag.EnvVar)
}

func TestParseTagWithQuotes(t *testing.T) {
	tag := `env=NAME,default="John Doe"`
	spec, err := ParseTag(tag)

	assert.NoError(t, err)
	assert.Equal(t, "NAME", spec.EnvVar)
	assert.Equal(t, "John Doe", spec.Default)
}

func TestParseTagRequired(t *testing.T) {
	tag := "env=SECRET,required"
	spec, err := ParseTag(tag)

	assert.NoError(t, err)
	assert.True(t, spec.Required)
	assert.Equal(t, "SECRET", spec.EnvVar)
}
