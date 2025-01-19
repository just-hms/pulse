package engine

import (
	"fmt"

	"github.com/just-hms/pulse/pkg/spimi/inverseindex"
)

type DocumentInfo struct {
	*inverseindex.Document
	Score float64
	ID    uint32
}

// Cmp implements heap.Orderable
func (doc *DocumentInfo) Cmp(other *DocumentInfo) bool {
	return doc.Score < other.Score
}

// String return the string representation of the DocInfo
func (doc *DocumentInfo) String() string {
	return fmt.Sprintf("{%f %s}", doc.Score, doc.No())
}
