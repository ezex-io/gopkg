package env_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/ezex-io/gopkg/env"
	"github.com/stretchr/testify/assert"
)

// TestGetEnvEmpty verifies ???
func TestGetEnvEmpty(t *testing.T) {
	assert.Equal(t, "", env.GetEnv[string]("MY_STRING"))
	assert.Equal(t, []string{}, env.GetEnv[[]string]("MY_STRING_LIST"))
}

// TestGetEnv verifies that environment variables are correctly parsed into supported types.
func TestGetEnv(t *testing.T) {
	t.Setenv("MY_INT", "1")
	t.Setenv("MY_BOOL", "true")
	t.Setenv("MY_FLOAT", "3.14")
	t.Setenv("MY_STRING", "str")
	t.Setenv("MY_STRING_LIST", "str1,str2")
	t.Setenv("MY_DURATION", "5m")

	assert.Equal(t, 1, env.GetEnv[int]("MY_INT"))
	assert.Equal(t, true, env.GetEnv[bool]("MY_BOOL"))
	assert.Equal(t, 3.14, env.GetEnv[float64]("MY_FLOAT"))
	assert.Equal(t, "str", env.GetEnv[string]("MY_STRING"))
	assert.Equal(t, []string{"str1", "str2"}, env.GetEnv[[]string]("MY_STRING_LIST"))
	assert.Equal(t, time.Minute*5, env.GetEnv[time.Duration]("MY_DURATION"))
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
	assert.Equal(t, []string{"str1", "str2"}, env.GetEnv[[]string]("MY_STRING_LIST", env.WithDefault("str1,str2")))
	assert.Equal(t, time.Second*5, env.GetEnv[time.Duration]("MY_DURATION", env.WithDefault("5s")))
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
	assert.Panics(t, func() {
		assert.Equal(t, "two seconds", env.GetEnv[time.Duration]("MY_DURATION"))
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
	assert.Panics(t, func() {
		assert.Equal(t, 2*time.Second, env.GetEnv[float64]("MY_FLOAT", env.WithDefault("2 seconds")))
	})
}

func TestLoadEnvsFromFileSuccess(t *testing.T) {
	tempDir := t.TempDir()
	envPath := filepath.Join(tempDir, ".env")

	err := os.WriteFile(envPath, []byte("FOO=bar"), 0o600)
	if err != nil {
		t.Fatalf("Failed to create test .env file: %v", err)
	}

	err = env.LoadEnvsFromFile(envPath)

	assert.NoError(t, err)
	assert.Equal(t, "bar", os.Getenv("FOO"))
}

func TestLoadEnvsFromFileFileNotFound(t *testing.T) {
	tempDir := t.TempDir()
	envPath := filepath.Join(tempDir, "file-not-exists.env")

	err := env.LoadEnvsFromFile(envPath)
	assert.Error(t, err)
}

func TestLoadEnvsFromFileEmptyPath(t *testing.T) {
	err := env.LoadEnvsFromFile()
	assert.Error(t, err)
}
