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
		wantPanic  bool
	}{
		{
			name:       "int with default",
			defaultVal: "42",
			expected:   42,
		},
		{
			name:       "bool true with default",
			defaultVal: "true",
			expected:   true,
		},
		{
			name:       "float with default",
			defaultVal: "3.14",
			expected:   3.14,
		},
		{
			name:       "string with default",
			defaultVal: "hello",
			expected:   "hello",
		},
		{
			name:      "unsupported type panics",
			expected:  struct{}{},
			wantPanic: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantPanic {
				assert.PanicsWithValue(t,
					"unsupported type: struct {}",
					func() { GetEnv[struct{}]("UNSUPPORTED") },
					"should panic for unsupported types")
				return
			}

			switch expected := tt.expected.(type) {
			case int:
				result := GetEnv[int]("MY_INT", WithDefault(tt.defaultVal))
				assert.Equal(t, expected, result)
			case bool:
				result := GetEnv[bool]("MY_BOOL", WithDefault(tt.defaultVal))
				assert.Equal(t, expected, result)
			case float64:
				result := GetEnv[float64]("MY_FLOAT", WithDefault(tt.defaultVal))
				assert.InDelta(t, expected, result, 0.0001)
			case string:
				result := GetEnv[string]("MY_STRING", WithDefault(tt.defaultVal))
				assert.Equal(t, expected, result)
			}
		})
	}
}
