package inverseindex

import (
	"io"
)

type Collection []Document

type Document struct {
	No  [8]rune
	Len uint32
}

func NewDocument(no string, _len int) Document {
	if len(no) > 8 {
		panic("dockNo cannot be bigger then 8")
	}
	return Document{
		// todo: fill
		No:  [8]rune{},
		Len: uint32(_len),
	}
}

func (c *Collection) Clear() {
	clear(*c)
	c = nil
}

func (c *Collection) Encode(w io.Writer) {
	// todo: write down everything with a fixed size DOCNO, don't write the ID is redundant
}

func DecodeDocument(ID uint32, rs io.ReadSeeker, d *Document) error {
	panic("unimplemented")
}
