package heap

import "container/heap"

// Define the custom type `integer` which implements the Orderable interface.
type integer int

// Orderable is an interface for types that can be compared.
type Orderable[T any] interface {
	Cmp(other T) bool
}

// Cmp defines how integers are compared. You can change the comparison logic as needed.
func (i integer) Cmp(other integer) bool {
	return i < other // Example: Min-Heap behavior
}

var _ heap.Interface = &container[integer]{}

// container is a generic heap container implementing `heap.Interface`.
type container[T Orderable[T]] struct {
	data []T
	cmp  func(a, b T) bool
}

// Len returns the number of elements in the heap.
func (h *container[T]) Len() int {
	return len(h.data)
}

// Less determines the ordering of elements. Uses the `cmp` function if provided.
func (h *container[T]) Less(i, j int) bool {
	if h.cmp != nil {
		return h.cmp(h.data[i], h.data[j])
	}
	return h.data[i].Cmp(h.data[j])
}

// Swap exchanges elements at indices i and j.
func (h *container[T]) Swap(i, j int) {
	h.data[i], h.data[j] = h.data[j], h.data[i]
}

// Push adds an element to the heap.
func (h *container[T]) Push(x any) {
	h.data = append(h.data, x.(T))
}

// Pop removes and returns the smallest element from the heap.
func (h *container[T]) Pop() any {
	old := h.data
	n := len(old)
	item := old[n-1]
	h.data = old[0 : n-1]
	return item
}

func (h *container[T]) Peek() (T, bool) {
	if h.Len() == 0 {
		var t T
		return t, false
	}
	return h.data[0], true
}
