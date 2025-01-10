package stats

import (
	"encoding/gob"
	"io"
)

type Stats struct {
	CollectionSize      uint32
	AverageDocumentSize float64
}

func (s *Stats) Dump(w io.Writer) error {
	enc := gob.NewEncoder(w)
	return enc.Encode(s)
}

func (s *Stats) Update(collectionSize uint32, averageDocumentSize float64) {
	s.AverageDocumentSize = (s.AverageDocumentSize*float64(s.CollectionSize) + averageDocumentSize*float64(collectionSize)) /
		float64(s.CollectionSize+collectionSize)
	s.CollectionSize += collectionSize
}

func Load(r io.Reader) (*Stats, error) {
	dec := gob.NewDecoder(r)
	c := &Stats{}
	err := dec.Decode(c)
	if err != nil {
		return c, err
	}
	return c, nil
}
