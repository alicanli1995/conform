package convert_test

import (
	"reflect"
	"testing"
	"time"

	"github.com/alicanli1995/conform/convert"
	"github.com/stretchr/testify/assert"
)

func TestConvertString(t *testing.T) {
	conv := convert.New()
	result, err := conv.Convert("hello", reflect.TypeOf(""), "")
	assert.NoError(t, err)
	assert.Equal(t, "hello", result)
}

func TestConvertInt(t *testing.T) {
	conv := convert.New()
	result, err := conv.Convert("123", reflect.TypeOf(0), "")
	assert.NoError(t, err)
	assert.Equal(t, 123, result)
}

func TestConvertBool(t *testing.T) {
	conv := convert.New()
	result, err := conv.Convert("true", reflect.TypeOf(false), "")
	assert.NoError(t, err)
	assert.Equal(t, true, result)
}

func TestConvertSlice(t *testing.T) {
	conv := convert.New()
	result, err := conv.Convert("1,2,3", reflect.TypeOf([]int{}), "")
	assert.NoError(t, err)
	assert.Equal(t, []int{1, 2, 3}, result)
}

func TestConvertTime(t *testing.T) {
	conv := convert.New()
	result, err := conv.Convert("2024-01-01T00:00:00Z", reflect.TypeOf(time.Time{}), time.RFC3339)
	assert.NoError(t, err)
	assert.IsType(t, time.Time{}, result)
}

func TestConvertDuration(t *testing.T) {
	conv := convert.New()
	result, err := conv.Convert("5s", reflect.TypeOf(time.Duration(0)), "")
	assert.NoError(t, err)
	assert.Equal(t, 5*time.Second, result)
}
