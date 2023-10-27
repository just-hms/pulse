package index

import "github.com/just-hms/pulse/pkg/entity"

type ChunkReader interface {
	Read() ([]*entity.Document, error)
}
