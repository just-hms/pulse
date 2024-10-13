package engine

import "github.com/just-hms/pulse/pkg/spimi/inverseindex"

type Result struct {
	inverseindex.Document
	Score float64
}
