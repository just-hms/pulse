package withkey

type WithKey[T any] struct {
	Key   string
	Value T
}
