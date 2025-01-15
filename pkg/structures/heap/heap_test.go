package heap_test

import (
	"slices"
	"testing"

	"github.com/just-hms/pulse/pkg/structures/heap"
	"github.com/stretchr/testify/require"
)

type Integer int

func (i Integer) Cmp(other Integer) bool { return i < other }

func TestPush(t *testing.T) {
	req := require.New(t)

	// Initialize the Heap
	h := &heap.Heap[Integer]{}
	h.Push(10, 5, 20)

	// Validate heap structure after pushes
	req.Equal(3, h.Size(), "heap should contain 3 elements")

	peek, ok := h.Peek()
	req.True(ok)
	req.Equal(Integer(5), peek, "top element should be the smallest (5)")
}

func TestPop(t *testing.T) {
	req := require.New(t)

	// Initialize the Heap
	h := &heap.Heap[Integer]{}
	h.Push(10, 5, 20)

	// Pop elements and validate order
	min, ok := h.Pop()
	req.True(ok)
	req.Equal(Integer(5), min, "first popped element should be 5")

	min, ok = h.Pop()
	req.True(ok)
	req.Equal(Integer(10), min, "second popped element should be 10")

	min, ok = h.Pop()
	req.True(ok)
	req.Equal(Integer(20), min, "third popped element should be 20")

	// Heap should now be empty
	_, ok = h.Pop()
	req.False(ok)
}

func TestPeek(t *testing.T) {
	req := require.New(t)

	// Initialize the Heap
	h := &heap.Heap[Integer]{}
	_, ok := h.Peek()
	req.False(ok)

	h.Push(10, 5, 20)

	// Peek at the top element without removing it
	min, ok := h.Peek()
	req.True(ok)
	req.Equal(Integer(5), min, "peeked element should be the smallest (5)")

	// Ensure size remains the same
	req.Equal(3, h.Size(), "heap size should remain 3 after peek")
}

func TestSize(t *testing.T) {
	require := require.New(t)

	// Initialize the Heap
	h := &heap.Heap[Integer]{}
	require.Equal(0, h.Size(), "new heap should be empty")

	h.Push(10, 5, 20)

	require.Equal(3, h.Size(), "heap should contain 3 elements after 3 pushes")

	// Pop one element and check size
	_, _ = h.Pop()
	require.Equal(2, h.Size(), "heap should contain 2 elements after one pop")
}

func TestValues(t *testing.T) {
	require := require.New(t)

	h := &heap.Heap[Integer]{}
	require.Equal(0, h.Size(), "new heap should be empty")

	h.Push(10, 5, 20, 100, -10, 200, 200)

	v := h.Values()

	require.True(slices.IsSorted(v), "heap is not sorted %v", v)
}

func TestDifferentOrder(t *testing.T) {
	require := require.New(t)

	h := (&heap.Heap[Integer]{}).WithOrder(func(a, b Integer) bool { return int(a) > int(b) })
	require.Equal(0, h.Size(), "new heap should be empty")

	h.Push(10, 5, 20, 100, -10, 200, 200)

	v := h.Values()
	slices.Reverse(v)

	require.True(slices.IsSorted(v), "heap is not sorted %v", v)
}

func TestMaxSize(t *testing.T) {
	require := require.New(t)

	h := (&heap.Heap[Integer]{Cap: 3}).WithOrder(func(a, b Integer) bool { return int(a) > int(b) })
	require.Equal(0, h.Size(), "new heap should be empty")

	h.Push(10, 5, 20, 100, -10, 200, 300)

	v := h.Values()
	slices.Reverse(v)
	require.Equal([]Integer{-10, 5, 10}, v)
}
