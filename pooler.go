package stdlib

import (
	"fmt"
	"sync"
)

type Pooler[T any] interface {
	Get() *T
	Clear()
	Size() int
}

type entityNode[T any] struct {
	Next  *entityNode[T]
	Value *T
}

type poolImpl[T any] struct {
	list     *entityNode[T]
	ptr      *entityNode[T]
	head     *entityNode[T]
	poolSize int
	mtx      sync.RWMutex
}

func (p *poolImpl[T]) Get() *T {
	if p.ptr == nil {
		return nil
	}
	p.mtx.Lock()
	defer p.mtx.Unlock()
	v := p.ptr.Value
	p.ptr = p.ptr.Next
	return v
}

func (p *poolImpl[T]) Size() int {
	p.mtx.RLock()
	defer p.mtx.RUnlock()
	return p.poolSize
}

func (p *poolImpl[T]) Clear() {
	p.mtx.Lock()
	defer p.mtx.Unlock()
	p.list = nil
	p.ptr = nil
	p.poolSize = 0
}

type PoolingOptions struct {
	PoolSize int
}

func NewPool[T any](
	initFunction func() *T,
	opt *PoolingOptions,
) (Pooler[T], error) {
	if opt == nil {
		return nil, fmt.Errorf("options cannot be nil")
	}
	if opt.PoolSize <= 0 {
		return nil, fmt.Errorf("pool size must be greater than 0")
	}
	p := &poolImpl[T]{}
	for i := 0; i < opt.PoolSize; i++ {
		ent := initFunction()
		node := &entityNode[T]{Value: ent}
		if i == 0 {
			p.list = node
			p.ptr = node
			p.head = node
		} else {
			p.ptr.Next = node
			p.ptr = node
		}
	}
	p.ptr.Next = p.head
	return p, nil
}
