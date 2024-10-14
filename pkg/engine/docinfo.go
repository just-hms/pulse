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

func (doc *DocInfo) Less(other *DocInfo) int {
	return int(doc.Score) - int(other.Score)
}

func (doc *DocInfo) String() string {
	return fmt.Sprintf("{%f %s}", doc.Score, string(doc.Document.No[:]))
}
