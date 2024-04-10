package spimi

type ChunkReader interface {
	Read() ([]Document, error)
}

type Document struct {
	No      string
	Content string
}
