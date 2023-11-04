package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStringBuilder(t *testing.T) {
	url, err := StringBuilder(
		"http://",
		"www.testtesttest.com",
		":",
		80,
		[]byte{'/'},
		4.32,
		[]rune{'/', 'a', 'b', 'c'},
	)
	assert.NoError(t, err)
	assert.Equal(t, "http://www.testtesttest.com:80/4.32/abc", url)
}

func TestStringBuilderNonSupportedType(t *testing.T) {
	_, err := StringBuilder(
		"http://",
		"www.testtesttest.com",
		":",
		80,
		[]byte{'/'},
		4.32,
		[]rune{'/', 'a', 'b', 'c'},
		rune(0x1F4A9),
		[]byte("test")[0],
		struct{}{},
	)
	assert.Error(t, err)
}

func TestStringBuilderAllTypes(t *testing.T) {
	url, err := StringBuilder(
		"http://",
		"www.testtesttest.com",
		":",
		80,
		[]byte{'/'},
		4.32,
		[]rune{'/', 'a', 'b', 'c'},
		"test",
		int8(1),
		int16(1),
		int32(1),
		int64(1),
		// uint8(1),
		uint16(1),
		uint32(1),
		uint64(1),
		float32(1.0),
		float64(1.0),
		true,
		false,
	)
	assert.NoError(t, err)
	assert.Equal(
		t,
		"http://www.testtesttest.com:80/4.32/abctest111111111truefalse",
		url,
	)
}
