package index

type With func(*Option)

type Option struct {
	ChunkSize int
}

func WithChunkSize(size int) func(*Option) {
	return func(o *Option) {
		o.ChunkSize = size
	}
}
