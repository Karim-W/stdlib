package stdlib

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTestPostionalArgs(t *testing.T) {
	str := "https://example.com/{server}/test/{test_name}"
	args := []string{"server1", "test1"}
	result := EmbedNamedPositionArgs(str, args...)
	assert.Equal(t, "https://example.com/server1/test/test1", result)
}

func TestInvokeGetRequest(t *testing.T) {
	url := "https://httpbin.org/get"
	cl, err := ClientProvider()
	assert.Nil(t, err)
	code, err := cl.Invoke(context.TODO(), "GET", url, nil, nil, nil)
	assert.Nil(t, err)
	assert.Equal(t, 200, code)
}
