package utils

import (
	"sync"
)

type HashSet[T comparable] struct {
	m  map[T]struct{}
	mu sync.RWMutex
}

func NewHashSet[T comparable](members ...T) *HashSet[T] {
	m := make(map[T]struct{})

	for _, member := range members {
		m[member] = struct{}{}
	}

	return &HashSet[T]{
		m: m,
	}
}

func (s *HashSet[T]) Add(value T) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.m[value] = struct{}{}
}

func (s *HashSet[T]) Contains(value T) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, ok := s.m[value]
	return ok
}

func (s *HashSet[T]) Remove(uid T) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.m, uid)
}

func (s *HashSet[T]) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.m = make(map[T]struct{})
}

func (s *HashSet[T]) Values() []T {
	s.mu.RLock()
	defer s.mu.RUnlock()

	values := make([]T, 0, len(s.m))
	for k := range s.m {
		values = append(values, k)
	}

	return values
}
