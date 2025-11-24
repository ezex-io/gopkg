package env

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type SupportedTypes interface {
	~string | ~int | ~float64 | ~bool | ~[]string | time.Duration
}

// Option defines a function type that modifies a string value,
// typically used for environment variable customization.
type Option func(value *string)

// WithDefault returns an Option that sets a default value
// if the environment variable is not set or is empty.
func WithDefault(defVal string) Option {
	return func(val *string) {
		if *val == "" {
			*val = defVal
		}
	}
}

// GetEnv retrieves an environment variable by key,
// applies the provided options, and converts it to the desired type T.
//
// Panics if the conversion fails or the type is unsupported.
func GetEnv[T SupportedTypes](key string, options ...Option) T {
	val := os.Getenv(key)
	for _, opt := range options {
		opt(&val)
	}

	var result T
	switch any(result).(type) {
	case int:
		v, err := strconv.Atoi(val)
		if err != nil {
			panic(fmt.Errorf("failed to convert %q to int: %w", val, err))
		}

		return any(v).(T)

	case bool:
		v, err := strconv.ParseBool(val)
		if err != nil {
			panic(fmt.Errorf("failed to convert %q to bool: %w", val, err))
		}

		return any(v).(T)

	case float64:
		v, err := strconv.ParseFloat(val, 64)
		if err != nil {
			panic(fmt.Errorf("failed to convert %q to float64: %w", val, err))
		}

		return any(v).(T)

	case string:
		return any(val).(T)

	case []string:
		if val == "" {
			return any([]string{}).(T)
		}
		parts := strings.Split(val, ",")

		return any(parts).(T)

	case time.Duration:
		dur, err := time.ParseDuration(val)
		if err != nil {
			panic(fmt.Errorf("failed to parse duration %q: %w", val, err))
		}

		return any(dur).(T)

	default:
		panic(fmt.Sprintf("unsupported type: %T", result))
	}
}

// LoadEnvsFromFile loads environment variables from the specified file(s).
// If a file is not found, it returns without an error.
func LoadEnvsFromFile(envFile ...string) error {
	return godotenv.Load(envFile...)
}
