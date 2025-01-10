package inverseindex

import (
	"bufio"
	"io"
)

type Collection struct {
	data            []Document
	AvgDocumentSize float64
}

func (c *Collection) Add(docs ...Document) {
	for _, d := range docs {
		c.data = append(c.data, d)

		// m_k := kth avg
		// a := general element
		// m_k = m_{k-1} + (a - m_{k-1}) / k
		c.AvgDocumentSize = c.AvgDocumentSize + float64(float64(d.Size)-c.AvgDocumentSize)/float64(len(c.data))
	}
}

func (c *Collection) Len() int {
	return len(c.data)
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
