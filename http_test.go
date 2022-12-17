package stdlib

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTestPostionalArgs(t *testing.T) {
	str := "https://example.com/{server}/test/{test_name}"
	args := []string{"server1", "test1"}
	result := EmbedNamedPositionArgs(str, args...)
	assert.Equal(t, "https://example.com/server1/test/test1", result)
}
