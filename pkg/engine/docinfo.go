package engine

import (
	"fmt"

	"github.com/just-hms/pulse/pkg/spimi/inverseindex"
)

type DocInfo struct {
	inverseindex.Document
	Score float64
	ID    uint32
}

// Cmp implements heap.Orderable.
func (doc *DocInfo) Cmp(other *DocInfo) bool {
	// todo: check this better
	return doc.Score < other.Score
}

func (doc *DocInfo) String() string {
	return fmt.Sprintf("{%f %s}", doc.Score, doc.No())
}
