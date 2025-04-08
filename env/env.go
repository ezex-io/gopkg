package env

import (
	"fmt"
	"strconv"
)

type Option func(value *string)

func WithDefault(defVal string) Option {
	return func(val *string) {
		if *val == "" {
			*val = defVal
		}
	}
}

func GetEnv[T any](key string, options ...Option) T {
	var val string
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
