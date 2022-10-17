package stdlib

import (
	"math/rand"
	"time"
)

const (
	strings string = "abcdefghijklmnopqrstuvwxyz001234567890123456789123456789"
)

func generateId(length int) string {
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, length)
	for i := range b {
		b[i] = strings[rand.Intn(len(strings))]
	}
	return string(b)
}
