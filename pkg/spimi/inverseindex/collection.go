package inverseindex

import (
	"bufio"
	"io"
)

type Collection []Document

func (c *Collection) Clear() {
	clear(*c)
	c = nil
}

func (c *Collection) Encode(w io.Writer) error {
	// todo: when to do the buffering?
	enc := bufio.NewWriter(w)
	defer enc.Flush()

	for _, doc := range *c {
		doc.Encode(enc)
	}
	return nil
}
