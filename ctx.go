package stdlib

import (
	"context"
	"time"
)

type Context struct {
	context.Context
	Res    chan interface{}
	Code   int
	values map[interface{}]interface{}
}

type Cancel func()

func NewContext() Context {
	return WrapContext(context.TODO())
}

func WrapContext(ctx context.Context) Context {
	return Context{
		Context: ctx,
		Res:     make(chan interface{}),
	}
}

func NewContextWithCancel() (Context, Cancel) {
	ctx, cancel := context.WithCancel(context.TODO())
	return WrapContext(ctx), Cancel(cancel)
}

func NewContextWithDeadline(deadline time.Time) (Context, Cancel) {
	ctx, cancel := context.WithDeadline(context.TODO(), deadline)
	return WrapContext(ctx), Cancel(cancel)
}

func (c *Context) Deadline() (deadline time.Time, ok bool) {
	return c.Context.Deadline()
}
func (c *Context) OnCancel() <-chan struct{} { return c.Context.Done() }
func (c *Context) Error() error              { return c.Context.Err() }

func (r *Context) CompletionHandler(code int, result interface{}) {
	r.Code = code
	r.Res <- result
}

func (r *Context) AddDeadline(deadline time.Time) Cancel {
	c, cancel := context.WithDeadline(r.Context, deadline)
	r.Context = c
	return Cancel(cancel)
}

func (r *Context) WithValue(key, val interface{}) {
	r.values[key] = val
	r.Context = context.WithValue(r.Context, key, val)
}

func (r *Context) Value(key string) interface{} {
	if v, ok := r.values[key]; ok {
		return v
	}
	return r.Context.Value(key)
}

func (r *Context) Cancel() {
	r.Context.Done()
}
