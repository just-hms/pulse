package stats

import (
	"encoding/gob"
	"io"
)

type IndexingSettings struct {
	Stemming, StopWordsRemoval, Compression bool
}
type Stats struct {
	IndexingSettings
	N   uint32  // collection size
	ADL float64 // average document length
}

func (s *Stats) Dump(w io.Writer) error {
	enc := gob.NewEncoder(w)
	return enc.Encode(s)
}

func (s *Stats) Update(collectionSize uint32, averageDocumentSize float64) {
	s.ADL = (s.ADL*float64(s.N) + averageDocumentSize*float64(collectionSize)) /
		float64(s.N+collectionSize)
	s.N += collectionSize
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
