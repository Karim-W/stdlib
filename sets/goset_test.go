package sets

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAddToIntSet(t *testing.T) {
	intSet := NewSet[int]()
	ok := intSet.Add(1)
	assert.True(t, ok)
	ok = intSet.Add(2)
	assert.True(t, ok)
	ok = intSet.Add(1)
	assert.False(t, ok)
	slice := intSet.ToSlice()
	assert.Equal(t, 2, len(slice))
	assert.Equal(t, 1, slice[0])
	assert.Equal(t, 2, slice[1])
}

func TestRemoveFromIntSet(t *testing.T) {
	intSet := NewSet[int]()
	ok := intSet.Add(1)
	assert.True(t, ok)
	ok = intSet.Add(2)
	assert.True(t, ok)
	ok = intSet.Remove(1)
	assert.True(t, ok)
	ok = intSet.Remove(1)
	assert.False(t, ok)
	slice := intSet.ToSlice()
	assert.Equal(t, 1, len(slice))
	assert.Equal(t, 2, slice[0])
}

func TestAddToStringSet(t *testing.T) {
	stringSet := NewSet[string]()
	ok := stringSet.Add("a")
	assert.True(t, ok)
	ok = stringSet.Add("b")
	assert.True(t, ok)
	ok = stringSet.Add("a")
	assert.False(t, ok)
	slice := stringSet.ToSlice()
	assert.Equal(t, 2, len(slice))
	assert.Contains(t, slice, "a")
	assert.Contains(t, slice, "b")
}

func TestSliceConversion(t *testing.T) {
	floatSet := NewSet[float64]()
	ok := floatSet.Add(1.0)
	assert.True(t, ok)
	ok = floatSet.Add(2.0)
	assert.True(t, ok)
	ok = floatSet.Add(1.2)
	assert.True(t, ok)
	slice := floatSet.ToSlice()
	assert.Equal(t, 3, len(slice))
	assert.Contains(t, slice, 1.0)
	assert.Contains(t, slice, 2.0)
	assert.Contains(t, slice, 1.2)
}
