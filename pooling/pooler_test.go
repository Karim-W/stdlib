package pooling

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPooler(t *testing.T) {
	// Createing a String Pool of 25
	p, err := NewPool(
		func() *string {
			str := "test"
			return &str
		},
		&PoolingOptions{
			PoolSize: 25,
		},
	)
	if err != nil {
		t.Error("must not return an error")
	}
	for i := 0; i < 100; i++ {
		ent := p.Get()
		if *ent != "test" {
			t.Error("MISMATCH")
		}
	}
}

func TestSinglePool(t *testing.T) {
	p, err := NewPool(
		func() *string {
			str := "test"
			return &str
		},
		&PoolingOptions{
			PoolSize: 1,
		},
	)
	size := p.Size()
	assert.Equal(t, size, 1)
	assert.Nil(t, err)
	ent := p.Get()
	assert.Equal(t, *ent, "test")
	ent = p.Get()
	assert.Equal(t, *ent, "test")
	ent = p.Get()
	assert.Equal(t, *ent, "test")
}

func TestPoolClear(t *testing.T) {
	p, err := NewPool(
		func() *string {
			str := "test"
			return &str
		},
		&PoolingOptions{
			PoolSize: 1,
		},
	)
	size := p.Size()
	assert.Equal(t, size, 1)
	assert.Nil(t, err)
	ent := p.Get()
	assert.Equal(t, *ent, "test")
	p.Clear()
	ent = p.Get()
	assert.Nil(t, ent)
}

func TestPoolSize(t *testing.T) {
	p, err := NewPool(
		func() *string {
			str := "test"
			return &str
		},
		&PoolingOptions{
			PoolSize: 25,
		},
	)
	assert.Nil(t, err)
	size := p.Size()
	assert.Equal(t, size, 25)
}
