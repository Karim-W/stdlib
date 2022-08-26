package stdlib

import (
	"testing"
)

func TestLogger(t *testing.T) {
	l := ConsoleLogger("[TEST SERVICE]")
	l.Info("hello")
	l.Error("hello")
}
