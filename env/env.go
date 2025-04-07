package env

import (
	"os"
	"strconv"
)

type Env interface {
	Get(key string, options ...Option) any
}

type Option func(*config)

type config struct {
	defaultValue any
	targetType   any
}

func WithDefault(value any) Option {
	return func(c *config) {
		c.defaultValue = value
	}
}

func WithType[T any](t T) Option {
	return func(c *config) {
		c.targetType = t
	}
}

var _ Env = envImpl{}

type envImpl struct{}

var EnvInstance Env = envImpl{}

func (e envImpl) Get(key string, options ...Option) any {
	val, exists := os.LookupEnv(key)
	cfg := &config{}

	for _, opt := range options {
		opt(cfg)
	}

	if !exists {
		return cfg.defaultValue
	}

	if cfg.targetType == nil {
		return val
	}

	switch cfg.targetType.(type) {
	case int:
		v, err := strconv.Atoi(val)
		if err != nil {
			panic(err)
		}
		return v
	case bool:
		v, err := strconv.ParseBool(val)
		if err != nil {
			panic(err)
		}
		return v
	case float64:
		v, err := strconv.ParseFloat(val, 64)
		if err != nil {
			panic(err)
		}
		return v
	case string:
		return val
	default:
		panic("unsupported type")
	}
}
