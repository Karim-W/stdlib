package stdlib

import (
	"math/rand"
	"time"
)

const (
	strs string = "abcdefghijklmnopqrstuvwxyz001234567890123456789123456789"
)

func generateId(length int) string {
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, length)
	for i := range b {
		b[i] = strs[rand.Intn(len(strs))]
	}
	return string(b)
}
