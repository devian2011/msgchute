package pagination

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetLimitsByPage(t *testing.T) {
	limit, offset := GetLimitsByPage(10, 14)
	assert.Equal(t, uint64(14), limit)
	assert.Equal(t, uint64(126), offset)
}

func TestGetPageCount(t *testing.T) {
	cnt := GetPageCount(10, 101)
	assert.Equal(t, 11, cnt)
}
