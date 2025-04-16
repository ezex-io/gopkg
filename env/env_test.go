package env_test

import (
	"os"
	"path/filepath"
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

func TestLoadEnvsFromFile(t *testing.T) {
	tempDir := t.TempDir()

	tests := []struct {
		name        string
		envContent  string
		envFileName string
		wantErr     bool
		setup       func() string
		cleanup     func()
	}{
		{
			name:        "successful load",
			envContent:  "TEST_KEY=test_value\nANOTHER_KEY=123",
			envFileName: ".env",
			wantErr:     false,
			setup: func() string {
				envPath := filepath.Join(tempDir, ".env")
				if err := os.WriteFile(envPath, []byte("TEST_KEY=test_value\nANOTHER_KEY=123"), 0o600); err != nil {
					t.Fatalf("Failed to create test .env file: %v", err)
				}

				return envPath
			},
			cleanup: func() {},
		},
		{
			name:        "file not found",
			envFileName: "nonexistent.env",
			wantErr:     true,
			setup: func() string {
				return filepath.Join(tempDir, "nonexistent.env")
			},
			cleanup: func() {},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			envPath := tt.setup()
			defer tt.cleanup()

			err := env.LoadEnvsFromFile(envPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadEnvsFromFile() error = %v, wantErr %v", err, tt.wantErr)

				return
			}

			if !tt.wantErr {
				if val := os.Getenv("TEST_KEY"); val != "test_value" {
					t.Errorf("TEST_KEY = %v, want test_value", val)
				}
				if val := os.Getenv("ANOTHER_KEY"); val != "123" {
					t.Errorf("ANOTHER_KEY = %v, want 123", val)
				}
			}
		})
	}
}
