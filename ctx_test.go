package stdlib

import (
	"testing"
	"time"
)

func TestContext(t *testing.T) {
	ctx := NewContext()
	go func() {
		time.Sleep(time.Second * 5)
		ctx.CompletionHandler(0, "hello")
	}()
	ctx.AddDeadline(time.Now().Add(time.Second * 1))
	select {
	case <-ctx.OnCancel():
		t.Log("done")
	case <-time.After(time.Second * 10):
		t.Error("timeout")
	case res := <-ctx.Res:
		t.Log(res)
	}
}
