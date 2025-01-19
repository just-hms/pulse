package set

import (
	"iter"
	"maps"
)

type Set[T comparable] map[T]struct{}

// Add adds an element to the set
func (s Set[T]) Add(value T) {
	s[value] = struct{}{}
}

// Remove removes an element from the set
func (s Set[T]) Remove(value T) {
	delete(s, value)
}

// Has checks if the set contains a given element
func (s Set[T]) Has(value T) bool {
	_, exists := s[value]
	return exists
}

// Values returns all elements in the set as a slice
func (s Set[T]) Values() iter.Seq[T] {
	return maps.Keys(s)
}
