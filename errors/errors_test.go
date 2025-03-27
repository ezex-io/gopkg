package errors

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	err := New(404, "not found")

	assert.Equal(t, 404, err.Code)
	assert.Equal(t, "not found", err.Message)
	assert.Empty(t, err.Meta)
	assert.Equal(t, "not found", err.Error())
}

func TestAddMeta_ValidPairs(t *testing.T) {
	err := New(400, "bad request").
		AddMeta("field", "email", "reason", "required")

	assert.Equal(t, "bad request", err.Message)
	assert.Equal(t, "email", err.Meta["field"])
	assert.Equal(t, "required", err.Meta["reason"])
	assert.Equal(t, 400, err.Code)
}

func TestAddMeta_InvalidPairs(t *testing.T) {
	err := New(400, "bad input").
		AddMeta("field", "email", "incomplete")

	assert.Contains(t, err.Meta, "error")
	assert.Equal(t, "invalid meta key/value args", err.Meta["error"])
}

func TestErrorMethod(t *testing.T) {
	err := New(500, "something broke")
	assert.Equal(t, "something broke", err.Error())
}
