package helper

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInArray_Strings(t *testing.T) {
	tests := []struct {
		name     string
		needle   string
		haystack []string
		want     bool
	}{
		{"found first", "a", []string{"a", "b", "c"}, true},
		{"found middle", "b", []string{"a", "b", "c"}, true},
		{"found last", "c", []string{"a", "b", "c"}, true},
		{"not found", "z", []string{"a", "b", "c"}, false},
		{"empty haystack", "a", []string{}, false},
		{"nil haystack", "a", nil, false},
		{"empty needle found", "", []string{"", "a"}, true},
		{"case sensitivity", "A", []string{"a"}, false},
		{"duplicates in haystack", "a", []string{"a", "a", "a"}, true},
		{"single element match", "a", []string{"a"}, true},
		{"single element no match", "b", []string{"a"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, InArray(tt.needle, tt.haystack))
		})
	}
}

func TestInArray_Ints(t *testing.T) {
	assert.True(t, InArray(2, []int{1, 2, 3}))
	assert.False(t, InArray(4, []int{1, 2, 3}))
	assert.False(t, InArray(1, []int{}))
	assert.False(t, InArray(0, nil))
}

type testPoint struct {
	X, Y int
}

func TestInArray_Structs(t *testing.T) {
	haystack := []testPoint{{1, 2}, {3, 4}}

	assert.True(t, InArray(testPoint{3, 4}, haystack))
	assert.False(t, InArray(testPoint{5, 6}, haystack))
}

func TestInArray_Pointers(t *testing.T) {
	// Pointers are compared by address, not by pointed-to value.
	a := &testPoint{1, 2}
	b := &testPoint{1, 2}

	assert.True(t, InArray(a, []*testPoint{a}))
	assert.False(t, InArray(b, []*testPoint{a}))
}
