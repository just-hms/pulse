package engine

import (
	"fmt"
	"maps"
	"slices"
	"strings"
)

type Metric uint8

const (
	TFIDF Metric = iota
	BM25
)

var (
	BM25_k1 = 1.3
	BM25_b  = 0.5
)

var metrics = map[string]Metric{
	"BM25":  BM25,
	"TFIDF": TFIDF,
}

var AllowedMetrics = slices.Collect(maps.Keys(metrics))

func ParseMetric(s string) (Metric, error) {
	s = strings.ToUpper(s)
	v, ok := metrics[s]
	if !ok {
		return 0, fmt.Errorf("%s is not a valid metric", s)
	}
	return v, nil
}
