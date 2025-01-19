package inverseindex

import (
	"bufio"
	"io"
)

// Collection contains the documents
type Collection struct {
	data            []Document
	AvgDocumentSize float64
}

func (c *Collection) Add(docs ...Document) {
	for _, d := range docs {
		c.data = append(c.data, d)

		// increamental mean
		// m_k := kth avg
		// a := general element
		// m_k = m_{k-1} + (a - m_{k-1}) / k
		c.AvgDocumentSize = c.AvgDocumentSize + float64(float64(d.Size)-c.AvgDocumentSize)/float64(len(c.data))
	}
}

// Len returns the number of documents contained in the collection
func (c *Collection) Len() int {
	return len(c.data)
}

// Clear clears the collection
func (c *Collection) Clear() {
	clear(c.data)
	c.data = nil
}

// Encode writes the collection into an io.Writer
func (c *Collection) Encode(w io.Writer) error {
	enc := bufio.NewWriter(w)
	defer enc.Flush()

	for _, doc := range c.data {
		if err := doc.Encode(enc); err != nil {
			return err
		}
	}
	return nil
}
