package main

import (
	"fmt"
	"log"
	"os"
	"reflect"
	"time"

	"github.com/alicanli1995/conform"
	"github.com/alicanli1995/conform/convert"
	"github.com/alicanli1995/conform/validate"
)

// CustomType demonstrates custom type conversion
type CustomType string

// CustomConfig with custom types and validators
type CustomConfig struct {
	CustomField CustomType    `conform:"env=CUSTOM_FIELD,default=test"`
	Duration    time.Duration `conform:"env=DURATION,default=5s"`
}

func main() {
	// Register custom converter
	converter := convert.New()
	converter.RegisterConverter(
		reflect.TypeOf(CustomType("")),
		func(s string) (interface{}, error) {
			return CustomType("custom_" + s), nil
		},
	)

	// Register custom validator
	validator := validate.New()
	validator.RegisterValidator("custom_check", func(value interface{}, params []string) error {
		str, ok := value.(string)
		if !ok {
			return fmt.Errorf("expected string")
		}
		if len(str) < 10 {
			return fmt.Errorf("value too short")
		}
		return nil
	})

	os.Setenv("CUSTOM_FIELD", "value")
	os.Setenv("DURATION", "10s")

	var cfg CustomConfig
	if err := conform.Load(&cfg,
		conform.WithConverter(converter),
		conform.WithValidator(validator),
	); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Custom Field: %s\n", cfg.CustomField)
	fmt.Printf("Duration: %s\n", cfg.Duration)
}
