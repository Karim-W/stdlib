package stdlib

import "testing"

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
