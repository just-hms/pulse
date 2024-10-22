package spimi

import "iter"

type ChunkReader interface {
	Read() iter.Seq2[[]Document, error]
	EOF() bool
}

type Document struct {
	No      string
	Content string
}
