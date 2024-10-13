package box_test

import (
	"testing"

	"github.com/just-hms/pulse/pkg/structures/box"

	"github.com/stretchr/testify/require"
)

func TestBoxAdd(t *testing.T) {
	t.Parallel()
	req := require.New(t)

	// Define a simple comparison function for integers
	cmp := func(a, b int) int { return a - b }

	// Create a new Box with integer type and maxSize set to 3
	b := box.NewBox(cmp, 3)

	// Add elements
	b.Add(1)
	b.Add(2)
	b.Add(3)

	// This should not alter the box since it's already at maxSize
	b.Add(4)

	// Verify size
	req.Equal(3, b.Size())

	// Verify content
	exp := []int{1, 2, 3}
	req.Equal(exp, b.Values())
}
