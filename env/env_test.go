package env_test

import (
	"os"
	"testing"

	"github.com/ezex-io/gopkg/env"
)

func TestEnv_Get(t *testing.T) {
	_ = os.Setenv("TEST_STRING", "hello")
	_ = os.Setenv("TEST_INT", "123")
	_ = os.Setenv("TEST_BOOL", "true")
	_ = os.Setenv("TEST_FLOAT", "3.14")

	tests := []struct {
		name     string
		key      string
		options  []env.Option
		expected any
	}{
		{
			name:     "string value",
			key:      "TEST_STRING",
			options:  []env.Option{env.WithType("")},
			expected: "hello",
		},
		{
			name:     "int value",
			key:      "TEST_INT",
			options:  []env.Option{env.WithType(0)},
			expected: 123,
		},
		{
			name:     "bool value",
			key:      "TEST_BOOL",
			options:  []env.Option{env.WithType(false)},
			expected: true,
		},
		{
			name:     "float value",
			key:      "TEST_FLOAT",
			options:  []env.Option{env.WithType(0.0)},
			expected: 3.14,
		},
		{
			name:     "no type casting",
			key:      "TEST_STRING",
			options:  nil,
			expected: "hello",
		},
		{
			name:     "default value used",
			key:      "NON_EXISTENT",
			options:  []env.Option{env.WithDefault("default!")},
			expected: "default!",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val := env.EnvInstance.Get(tt.key, tt.options...)
			if val != tt.expected {
				t.Errorf("expected %v (%T), got %v (%T)", tt.expected, tt.expected, val, val)
			}
		})
	}
}
