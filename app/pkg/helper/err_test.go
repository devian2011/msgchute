package helper

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrMapToError_NilMap(t *testing.T) {
	err := ErrMapToError(nil)

	assert.NotNil(t, err)
	assert.Equal(t, "", err.Error(), "nil map produces an error with an empty message")
}

func TestErrMapToError_EmptyMap(t *testing.T) {
	err := ErrMapToError(map[string]error{})

	assert.NotNil(t, err)
	assert.Equal(t, "", err.Error())
}

func TestErrMapToError_SingleEntry(t *testing.T) {
	err := ErrMapToError(map[string]error{
		"transport": errors.New("unknown transport: push"),
	})

	assert.NotNil(t, err)
	assert.Equal(t, "transport: unknown transport: push\n", err.Error())
}

func TestErrMapToError_MultipleEntries(t *testing.T) {
	// Map iteration order is not deterministic, so we validate line-by-line
	// rather than comparing the whole string.
	input := map[string]error{
		"transport": errors.New("unknown transport: push"),
		"message":   errors.New("empty message"),
		"user_id":   errors.New("required"),
	}

	err := ErrMapToError(input)
	assert.NotNil(t, err)

	raw := err.Error()
	assert.True(t, strings.HasSuffix(raw, "\n"), "each entry is terminated with a newline")

	lines := strings.Split(strings.TrimRight(raw, "\n"), "\n")
	assert.Len(t, lines, len(input), "one line per entry")

	for _, line := range lines {
		parts := strings.SplitN(line, ": ", 2)
		assert.Len(t, parts, 2, "each line must be 'key: value'")

		key, value := parts[0], parts[1]
		expected, ok := input[key]
		assert.True(t, ok, "unexpected key %q", key)
		assert.Equal(t, expected.Error(), value)
	}
}

func TestErrMapToError_ValueWithColon(t *testing.T) {
	// "unknown transport: push" already contains ": " — SplitN with limit 2
	// must keep the value intact.
	err := ErrMapToError(map[string]error{
		"transport": errors.New("unknown transport: push"),
	})

	assert.Equal(t, "transport: unknown transport: push\n", err.Error())
}
