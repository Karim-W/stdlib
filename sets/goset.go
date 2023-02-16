package sets

import "sync"

type Set[T comparable] interface {
	// Add adds an element to the set.
	Add(elem T) bool
	// Remove removes an element from the set.
	Remove(elem T) bool
	// Creates a unique slice new set with the elements
	ToSlice() []T
}

type _Set[T comparable] struct {
	s   map[T]bool
	mtx sync.RWMutex
}

// NewSet creates a new set.
func NewSet[T comparable]() Set[T] {
	return &_Set[T]{
		s:   make(map[T]bool),
		mtx: sync.RWMutex{},
	}
}

// Add() adds an element to the set.
// Returns true if the element was added, false if it was already present.
func (s *_Set[T]) Add(elem T) bool {
	_, ok := s.s[elem]
	if ok {
		return false
	}
	s.mtx.Lock()
	s.s[elem] = true
	s.mtx.Unlock()
	return true
}

func (s *_Set[T]) Remove(elem T) bool {
	_, ok := s.s[elem]
	if !ok {
		return false
	}
	s.mtx.Lock()
	delete(s.s, elem)
	s.mtx.Unlock()
	return true
}

func (s *_Set[T]) ToSlice() []T {
	res := make([]T, 0, len(s.s))
	s.mtx.RLock()
	for k := range s.s {
		res = append(res, k)
	}
	s.mtx.RUnlock()
	return res
}
