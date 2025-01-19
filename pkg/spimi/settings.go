package spimi

import (
	"github.com/just-hms/pulse/pkg/preprocess"
)

// IndexingSettings are the settings used to build the index
type IndexingSettings struct {
	PreprocessSettings preprocess.Settings
	Compression        bool
	MemoryThresholdMB  int
	NumWorkers         int
}
