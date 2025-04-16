package env

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Option defines a function type that modifies a string value, typically used for environment variable customization.
type Option func(value *string)

// WithDefault returns an Option that sets a default value if the environment variable is empty.
func WithDefault(defVal string) Option {
	return func(val *string) {
		if *val == "" {
			*val = defVal
		}
	}
}

// GetEnv retrieves an environment variable by key,
// applies optional modifications, and converts it to the desired type T.
// Supports types: int, bool, float64, and string.
// Panics if conversion fails or if the type is unsupported.
func GetEnv[T any](key string, options ...Option) T {
	val := os.Getenv(key)
	for _, opt := range options {
		opt(&val)
	}

	var result T
	switch any(result).(type) {
	case int:
		v, err := strconv.Atoi(val)
		if err != nil {
			panic(err)
		}

		return any(v).(T)
	case bool:
		v, err := strconv.ParseBool(val)
		if err != nil {
			panic(err)
		}

		return any(v).(T)
	case float64:
		v, err := strconv.ParseFloat(val, 64)
		if err != nil {
			panic(err)
		}

		return any(v).(T)
	case string:
		return any(val).(T)
	default:
		panic(fmt.Sprintf("unsupported type: %T", result))
	}
}

func LoadEnvsFromFile(envFile string) error {
	if err := godotenv.Load(envFile); err != nil {
		return fmt.Errorf("failed to load environment variables from file %s: %w", envFile, err)
	}

	return nil
}
