package utils

import (
	"testing"
	"unicode"

	"github.com/stretchr/testify/assert"
)

func TestGenerateRandomCode(t *testing.T) {
	t.Run("DefaultDigitsOnly", func(t *testing.T) {
		code, err := GenerateRandomCode(6, "")
		assert.NoError(t, err)
		assert.Equal(t, 6, len(code), "code length should be 6")

		for _, ch := range code {
			assert.True(t, unicode.IsDigit(ch), "each character should be a digit")
		}
	})

	t.Run("AlphaNumeric", func(t *testing.T) {
		code, err := GenerateRandomCode(10, AlphaNumeric)
		assert.NoError(t, err)
		assert.Equal(t, 10, len(code))

		for _, ch := range code {
			assert.Contains(t, AlphaNumeric, string(ch), "character should be in alphanumeric charset")
		}
	})

	t.Run("CustomCharset", func(t *testing.T) {
		charset := "ABC123"
		code, err := GenerateRandomCode(5, charset)
		assert.NoError(t, err)
		assert.Equal(t, 5, len(code))

		for _, ch := range code {
			assert.Contains(t, charset, string(ch), "character should be in custom charset")
		}
	})

	t.Run("ZeroLength", func(t *testing.T) {
		code, err := GenerateRandomCode(0, "")
		assert.Error(t, err)
		assert.Empty(t, code)
	})

	t.Run("Uniqueness", func(t *testing.T) {
		code1, err1 := GenerateRandomCode(8, Digits)
		code2, err2 := GenerateRandomCode(8, Digits)
		assert.NoError(t, err1)
		assert.NoError(t, err2)

		// Rare case of equality is accepted but generally they should differ
		assert.NotEqual(t, code1, code2, "codes should be different (usually)")
	})
}
