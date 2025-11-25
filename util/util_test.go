package util

import (
	"testing"
	"unicode"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateRandomCode(t *testing.T) {
	t.Run("DefaultDigitsOnly", func(t *testing.T) {
		code, err := GenerateRandomCode(6, "")
		require.NoError(t, err)
		assert.Len(t, code, 6, "code length should be 6")

		for _, ch := range code {
			assert.True(t, unicode.IsDigit(ch), "each character should be a digit")
		}
	})

	t.Run("AlphaNumeric", func(t *testing.T) {
		code, err := GenerateRandomCode(10, AlphaNumeric)
		require.NoError(t, err)
		assert.Len(t, code, 10)

		for _, ch := range code {
			assert.Contains(t, AlphaNumeric, string(ch), "character should be in alphanumeric charset")
		}
	})

	t.Run("CustomCharset", func(t *testing.T) {
		charset := "ABC123"
		code, err := GenerateRandomCode(5, charset)
		require.NoError(t, err)
		assert.Len(t, code, 5)

		for _, ch := range code {
			assert.Contains(t, charset, string(ch), "character should be in custom charset")
		}
	})

	t.Run("ZeroLength", func(t *testing.T) {
		code, err := GenerateRandomCode(0, "")
		require.Error(t, err)
		assert.Empty(t, code)
	})

	t.Run("Uniqueness", func(t *testing.T) {
		code1, err1 := GenerateRandomCode(8, Digits)
		code2, err2 := GenerateRandomCode(8, Digits)
		require.NoError(t, err1)
		require.NoError(t, err2)

		// Rare case of equality is accepted but generally they should differ
		assert.NotEqual(t, code1, code2, "codes should be different (usually)")
	})
}
