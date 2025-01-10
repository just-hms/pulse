package heap

import "errors"

type Orderable[T any] interface {
	Cmp(other T) int
}

type Heap[T Orderable[T]] struct {
	content []T
	cmp     func(a, b T) int
}

// NewDifferentOrder initializes a new heap with a comparator function.
func NewDifferentOrder[T Orderable[T]](cmp func(a, b T) int) Heap[T] {
	return Heap[T]{
		content: make([]T, 0),
		cmp:     cmp,
	}
}

// NewDifferentOrder initializes a new heap with a comparator function.
func New[T Orderable[T]]() Heap[T] {
	return Heap[T]{
		content: make([]T, 0),
	}
}

// Add inserts elements into the heap and maintains the heap property.
func (h *Heap[T]) Add(ts ...T) {
	if h.cmp == nil {
		h.cmp = func(a, b T) int { return a.Cmp(b) }
	}

	for _, t := range ts {
		h.content = append(h.content, t)
		h.upHeap(len(h.content) - 1)
	}
}

// Resize changes the capacity of the underlying array.
// If the new size is smaller, elements are truncated.
func (h *Heap[T]) Resize(size int) {
	h.content = h.content[:size]
}

// SortedValues returns the values in the heap in sorted order.
// It does so by creating a temporary copy of the heap and popping elements.
func (h *Heap[T]) Values() []T {
	// Create a copy of the heap content
	tempHeap := Heap[T]{
		content: make([]T, len(h.content)),
		cmp:     h.cmp,
	}
	copy(tempHeap.content, h.content)

	// Down-heap all elements to maintain heap structure in the temporary heap
	for i := len(tempHeap.content)/2 - 1; i >= 0; i-- {
		tempHeap.downHeap(i)
	}

	// Pop elements from the temporary heap to get them in sorted order
	result := make([]T, 0, len(h.content))
	for tempHeap.Size() > 0 {
		val, _ := tempHeap.Pop()
		result = append(result, val)
	}
	return result
}

// Size returns the number of elements in the heap.
func (h *Heap[T]) Size() int {
	return len(h.content)
}

// Peek returns the top element of the heap without removing it.
// Assumes the heap is not empty.
func (h *Heap[T]) Peek() (T, bool) {
	if len(h.content) == 0 {
		var zero T
		return zero, false
	}
	return h.content[0], true
}

// upHeap maintains the heap property by moving an element up.
func (h *Heap[T]) upHeap(index int) {
	for index > 0 {
		parent := (index - 1) / 2
		if h.cmp(h.content[index], h.content[parent]) >= 0 {
			break
		}
		h.swap(index, parent)
		index = parent
	}
}

// downHeap maintains the heap property by moving an element down.
func (h *Heap[T]) downHeap(index int) {
	lastIndex := len(h.content) - 1
	for {
		leftChild := 2*index + 1
		rightChild := 2*index + 2
		smallest := index

		if leftChild <= lastIndex && h.cmp(h.content[leftChild], h.content[smallest]) < 0 {
			smallest = leftChild
		}
		if rightChild <= lastIndex && h.cmp(h.content[rightChild], h.content[smallest]) < 0 {
			smallest = rightChild
		}
		if smallest == index {
			break
		}
		h.swap(index, smallest)
		index = smallest
	}
}

// swap swaps two elements in the content slice.
func (h *Heap[T]) swap(i, j int) {
	h.content[i], h.content[j] = h.content[j], h.content[i]
}

func (h *Heap[T]) Pop() (T, error) {
	if len(h.content) == 0 {
		var zero T
		return zero, errors.New("no element")
	}
	root := h.content[0]
	h.content[0] = h.content[len(h.content)-1]
	h.content = h.content[:len(h.content)-1]
	h.downHeap(0)
	return root, nil
}
