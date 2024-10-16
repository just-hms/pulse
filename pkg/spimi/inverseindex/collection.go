package inverseindex

import (
	"bufio"
	"io"
)

// todo: add avg document size
type Collection struct {
	data      []Document
	totalSize uint32
}

func (c *Collection) Add(docs ...Document) {
	for _, d := range docs {
		c.data = append(c.data, d)
		c.totalSize += d.Size
	}
}

func (c *Collection) Len() int {
	return len(c.data)
}

func (c *Collection) AverageDocumentSize() float64 {
	return float64(c.totalSize) / float64(len(c.data))
}

func (c *Collection) Clear() {
	clear(c.data)
	c.data = nil
}

func (c *Collection) Encode(w io.Writer) error {
	enc := bufio.NewWriter(w)
	defer enc.Flush()

	for _, doc := range c.data {
		doc.Encode(enc)
	}
	return nil
}
