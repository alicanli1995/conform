package validate_test

import (
	"testing"

	"github.com/alicanli1995/conform/parser"
	"github.com/alicanli1995/conform/validate"
	"github.com/stretchr/testify/assert"
)

func TestValidateEmail(t *testing.T) {
	v := validate.New()
	rules := []parser.Validator{{Name: "email"}}
	err := v.Validate("user@example.com", rules)
	assert.NoError(t, err)
}

func TestValidateEmailInvalid(t *testing.T) {
	v := validate.New()
	rules := []parser.Validator{{Name: "email"}}
	err := v.Validate("invalid-email", rules)
	assert.Error(t, err)
}

func TestValidateMin(t *testing.T) {
	v := validate.New()
	rules := []parser.Validator{{Name: "min", Params: []string{"5"}}}
	err := v.Validate("hello world", rules)
	assert.NoError(t, err)
}

func TestValidateGte(t *testing.T) {
	v := validate.New()
	rules := []parser.Validator{{Name: "gte", Params: []string{"10"}}}
	err := v.Validate(15, rules)
	assert.NoError(t, err)
}

func TestValidateLte(t *testing.T) {
	v := validate.New()
	rules := []parser.Validator{{Name: "lte", Params: []string{"100"}}}
	err := v.Validate(50, rules)
	assert.NoError(t, err)
}

func TestValidateOneOf(t *testing.T) {
	v := validate.New()
	rules := []parser.Validator{{Name: "oneof", Params: []string{"red", "green", "blue"}}}
	err := v.Validate("red", rules)
	assert.NoError(t, err)
}

func TestValidateRegex(t *testing.T) {
	v := validate.New()
	rules := []parser.Validator{{Name: "regex", Params: []string{"^[0-9]+$"}}}
	err := v.Validate("12345", rules)
	assert.NoError(t, err)
}
