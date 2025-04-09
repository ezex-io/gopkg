package env_test

import (
	"testing"

	"github.com/ezex-io/gopkg/env"
	"github.com/stretchr/testify/assert"
)

// TestGetEnv verifies that environment variables are correctly parsed into supported types.
func TestGetEnv(t *testing.T) {
	t.Setenv("MY_INT", "1")
	t.Setenv("MY_BOOL", "true")
	t.Setenv("MY_FLOAT", "3.14")
	t.Setenv("MY_STRING", "str")

	assert.Equal(t, 1, env.GetEnv[int]("MY_INT"))
	assert.Equal(t, true, env.GetEnv[bool]("MY_BOOL"))
	assert.Equal(t, 3.14, env.GetEnv[float64]("MY_FLOAT"))
	assert.Equal(t, "str", env.GetEnv[string]("MY_STRING"))
}

// TestGetEnvWithDefault verifies that default values are used when environment variables are not set.
func TestGetEnvWithDefault(t *testing.T) {
	assert.Equal(t, 1, env.GetEnv[int]("MY_INT", env.WithDefault("1")))
	assert.Equal(t, false, env.GetEnv[bool]("MY_BOOL", env.WithDefault("false")))
	assert.Equal(t, true, env.GetEnv[bool]("MY_BOOL", env.WithDefault("true")))
	assert.Equal(t, false, env.GetEnv[bool]("MY_BOOL", env.WithDefault("0")))
	assert.Equal(t, true, env.GetEnv[bool]("MY_BOOL", env.WithDefault("1")))
	assert.Equal(t, 3.14, env.GetEnv[float64]("MY_FLOAT", env.WithDefault("3.14")))
	assert.Equal(t, "str", env.GetEnv[string]("MY_STRING", env.WithDefault("str")))
}

// TestGetEnvNotSet ensures that calling GetEnv without a default on an unset variable panics.
func TestGetEnvNotSet(t *testing.T) {
	assert.Panics(t, func() {
		assert.Equal(t, 1, env.GetEnv[int]("MY_INT"))
	})
	assert.Panics(t, func() {
		assert.Equal(t, true, env.GetEnv[bool]("MY_BOOL"))
	})
	assert.Panics(t, func() {
		assert.Equal(t, 3.14, env.GetEnv[float64]("MY_FLOAT"))
	})
}

// TestGetEnvWrongType checks that GetEnv panics when default values cannot be parsed into the desired type.
func TestGetEnvWrongType(t *testing.T) {
	assert.Panics(t, func() {
		assert.Equal(t, 1, env.GetEnv[int]("MY_INT", env.WithDefault("one")))
	})
	assert.Panics(t, func() {
		assert.Equal(t, true, env.GetEnv[bool]("MY_BOOL", env.WithDefault("ok")))
	})
	assert.Panics(t, func() {
		assert.Equal(t, 3.14, env.GetEnv[float64]("MY_FLOAT", env.WithDefault("pi")))
	})
}

// TestGetEnvUnsupported ensures that GetEnv panics when an unsupported type is requested.
func TestGetEnvUnsupported(t *testing.T) {
	assert.Panics(t, func() {
		assert.Equal(t, 1, env.GetEnv[[]int]("MY_INT_ARRAY", env.WithDefault("[1]")))
	})
}
