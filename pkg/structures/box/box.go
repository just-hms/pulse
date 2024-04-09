package box

type CmpFunc[T any] func(a, b T) int

type Box[T any] interface {
	Add(el ...T)
	Size() int
	List() []T
}

var _ Box[int] = &box[int]{}

func NewBox[T any](cmp CmpFunc[T], maxSize int) Box[T] {
	return &box[T]{
		cmp:     cmp,
		maxSize: maxSize,
	}
}

type box[T any] struct {
	content []T
	maxSize int
	cmp     func(a, b T) int
}

// TODO: make it variadic

// Add add a element to a box
func (b *box[T]) Add(els ...T) {
	for _, el := range els {
		if len(b.content) < b.maxSize {
			b.content = append(b.content, el)
			continue
		}

		maxIdx := 0
		for i := 1; i < len(b.content); i++ {
			if b.cmp(b.content[i], b.content[maxIdx]) > 0 {
				maxIdx = i
			}
		}
		// if the new element is less then the max
		if b.cmp(el, b.content[maxIdx]) < 0 {
			b.content[maxIdx] = el
		}
	}
}

// List given a box returns a list
func (b *box[T]) List() []T {
	return b.content
}

// Size returns the box's size
func (b *box[T]) Size() int {
	return len(b.content)
}
