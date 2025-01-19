package spimi

import "iter"

type Chunk = []Document

// ChunkReader interface used by spimi to read a dataset
//
// implement this interface to read a different dataset
type ChunkReader interface {
	Read() iter.Seq2[Chunk, error]
	EOF() bool
}
