package heap

type Heap[T Orderable[T]] struct {
	c   container[T]
	Cap int
}

// Set the cmp function of the heap and returns its reference
func (h *Heap[T]) WithOrder(cmp func(a, b T) bool) *Heap[T] {
	h.c.cmp = cmp
	return h
}

// Len returns the number of elements in the heap.
func (h *Heap[T]) Len() int {
	return h.c.Len()
}

// Pop removes and returns the smallest/largest element from the heap (interface method).
func (h *Heap[T]) Pop() (T, bool) {
	if h.c.Len() == 0 {
		var t T
		return t, false
	}
	v := h.c.Pop()
	return v.(T), true
}

// Insert adds a new element to the fixed-size heap
func (h *Heap[T]) Push(vals ...T) {
	for _, v := range vals {
		if h.Cap == 0 || h.Len() < h.Cap {
			h.c.Push(v)
			continue
		}
		if peek, ok := h.Peek(); ok && h.c.cmp(v, peek) {
			h.Pop()
			h.c.Push(v)
		}
	}
}

// Peek returns the top element of the heap (smallest/largest).
func (h *Heap[T]) Peek() (T, bool) {
	return h.c.Peek()
}

// Values returns all elements in the heap by copying and popping each value.
// The original heap remains unchanged.
func (h *Heap[T]) Values() []T {
	if h.c.Len() == 0 {
		return []T{}
	}

	// Clone the container to avoid modifying the original heap
	tempContainer := &container[T]{
		data: append([]T{}, h.c.data...), // Copy the data
		cmp:  h.c.cmp,                    // Use the same comparison function
	}

	result := make([]T, 0, tempContainer.Len())
	for tempContainer.Len() > 0 {
		result = append(result, tempContainer.Pop().(T))
	}
	return result
}
