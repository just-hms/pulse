package withkey

// WithKey is a Pair<string,T>
type WithKey[T any] struct {
	Key   string
	Value T
}
