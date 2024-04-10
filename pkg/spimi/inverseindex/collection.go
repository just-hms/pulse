package inverseindex

import (
	"encoding/gob"
	"io"
)

type Collection []Document

type Document struct {
	ID  uint32
	No  string
	Len int
}

func (c *Collection) Clear() {
	*c = make(Collection, 0)
	c = nil
}

func (c *Collection) Encode(w io.Writer) {
	enc := gob.NewEncoder(w)
	for _, doc := range *c {
		enc.Encode(doc)
	}
}
