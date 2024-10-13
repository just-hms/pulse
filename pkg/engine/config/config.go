package config

import (
	"encoding/gob"
	"io"
)

type Config struct {
	Partitions []uint32
}

func (c *Config) Dump(w io.Writer) error {
	enc := gob.NewEncoder(w)
	return enc.Encode(c)
}

func Load(r io.Reader) (*Config, error) {
	dec := gob.NewDecoder(r)
	c := &Config{}
	err := dec.Decode(c)
	if err != nil {
		return c, err
	}
	return c, nil
}
