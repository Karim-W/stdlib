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
	list     *entityNode[T] //circullar linked list
	ptr      *entityNode[T] //pointer to current node
	head     *entityNode[T] //pointer to head node
	poolSize int
	mtx      sync.RWMutex
}

func (p *poolImpl[T]) Get() *T {
	if p.ptr == nil {
		return nil
	}
	p.mtx.Lock()
	defer p.mtx.Unlock()
	ent := p.ptr.Value
	p.ptr = p.ptr.Next
	return ent
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
		if p.list == nil {
			p.list = node
			p.ptr = node
			p.head = node
		} else {
			p.list.Next = node
			p.list = node
		}
	}
	p.list.Next = p.head
	p.poolSize = opt.PoolSize
	return p, nil
}
