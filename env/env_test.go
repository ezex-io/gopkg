package env

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetEnv(t *testing.T) {
	tests := []struct {
		name       string
		defaultVal string
		expected   any
		resultType string // just for clarity in logs
		assertFunc func(t *testing.T, actual, expected any)
	}{
		{
			name:       "int with default",
			defaultVal: "42",
			expected:   42,
			resultType: "int",
			assertFunc: func(t *testing.T, actual, expected any) {
				assert.Equal(t, expected.(int), actual.(int))
			},
		},
		{
			name:       "bool true with default",
			defaultVal: "true",
			expected:   true,
			resultType: "bool",
			assertFunc: func(t *testing.T, actual, expected any) {
				assert.Equal(t, expected.(bool), actual.(bool))
			},
		},
		{
			name:       "float with default",
			defaultVal: "3.14",
			expected:   3.14,
			resultType: "float64",
			assertFunc: func(t *testing.T, actual, expected any) {
				assert.InDelta(t, expected.(float64), actual.(float64), 0.0001)
			},
		},
		{
			name:       "string with default",
			defaultVal: "hello",
			expected:   "hello",
			resultType: "string",
			assertFunc: func(t *testing.T, actual, expected any) {
				assert.Equal(t, expected.(string), actual.(string))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			switch tt.expected.(type) {
			case int:
				result := GetEnv[int]("MY_INT", WithDefault(tt.defaultVal))
				tt.assertFunc(t, result, tt.expected)
			case bool:
				result := GetEnv[bool]("MY_BOOL", WithDefault(tt.defaultVal))
				tt.assertFunc(t, result, tt.expected)
			case float64:
				result := GetEnv[float64]("MY_FLOAT", WithDefault(tt.defaultVal))
				tt.assertFunc(t, result, tt.expected)
			case string:
				result := GetEnv[string]("MY_STRING", WithDefault(tt.defaultVal))
				tt.assertFunc(t, result, tt.expected)
			default:
				t.Fatalf("unsupported test type: %T", tt.expected)
			}
		})
	}
}
