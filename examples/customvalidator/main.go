package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"unicode"

	"github.com/alicanli1995/conform"
)

// Config demonstrates custom validators
type Config struct {
	Password string `conform:"env=PASSWORD,required"`
	Email    string `conform:"env=EMAIL,required,validate=email"`
}

func main() {
	// Register custom validator
	conform.RegisterValidator("strong_password", func(value interface{}, params []string) error {
		str, ok := value.(string)
		if !ok {
			return errors.New("expected string")
		}

		if len(str) < 12 {
			return errors.New("password must be at least 12 characters")
		}

		hasUpper := false
		hasLower := false
		hasDigit := false
		hasSpecial := false

		for _, ch := range str {
			if unicode.IsUpper(ch) {
				hasUpper = true
			}
			if unicode.IsLower(ch) {
				hasLower = true
			}
			if unicode.IsDigit(ch) {
				hasDigit = true
			}
			if !unicode.IsLetter(ch) && !unicode.IsDigit(ch) && !unicode.IsSpace(ch) {
				hasSpecial = true
			}
		}

		var missing []string
		if !hasUpper {
			missing = append(missing, "uppercase letter")
		}
		if !hasLower {
			missing = append(missing, "lowercase letter")
		}
		if !hasDigit {
			missing = append(missing, "digit")
		}
		if !hasSpecial {
			missing = append(missing, "special character")
		}

		if len(missing) > 0 {
			return fmt.Errorf("password must contain: %s", strings.Join(missing, ", "))
		}

		return nil
	})

	os.Setenv("EMAIL", "user@example.com")
	os.Setenv("PASSWORD", "WeakPass123") // This will fail

	cfg, err := conform.LoadGeneric[Config](conform.FromEnv())
	if err != nil {
		fmt.Println("Error (as expected):")
		fmt.Println(err)
		return
	}

	fmt.Printf("Config loaded: %+v\n", cfg)
}
