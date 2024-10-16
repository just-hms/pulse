package stats

import (
	"encoding/gob"
	"io"
)

type Stats struct {
	CollectionSize uint32
}

func (c *Stats) Dump(w io.Writer) error {
	enc := gob.NewEncoder(w)
	return enc.Encode(c)
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
