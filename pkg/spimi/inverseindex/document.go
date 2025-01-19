package inverseindex

import (
	"bytes"
	"encoding/binary"
	"io"
	"unsafe"
)

type Document struct {
	no   [8]byte // the maximum document number is set to be 8 (this is needed for O(1) access to document in disk)
	Size uint32
}

const DOC_SIZE = unsafe.Sizeof(Document{})

// NewDocument creates a new document
func NewDocument(no string, size int) Document {
	if len(no) > 8 {
		panic("dockNo cannot be bigger then 8")
	}
	d := Document{
		Size: uint32(size),
		no:   [8]byte{},
	}
	copy(d.no[:], no)
	return d
}

// No return the document number trimming out every 0's byte
func (doc *Document) No() string {
	return string(bytes.Trim(doc.no[:], "\x00"))
}

// Encode encodes a document into an io.Writer
func (doc *Document) Encode(w io.Writer) error {
	if err := binary.Write(w, binary.LittleEndian, doc.Size); err != nil {
		return err
	}
	if err := binary.Write(w, binary.LittleEndian, doc.no); err != nil {
		return err
	}
	return nil
}

// Decode decodes a document given its ID and a io.ReaderAt
func (doc *Document) Decode(ID uint32, rs io.ReaderAt) error {
	offset := int64(DOC_SIZE) * int64(ID)

	var buf [4]byte

	if _, err := rs.ReadAt(buf[:], offset+0); err != nil {
		return err
	}
	doc.Size = binary.LittleEndian.Uint32(buf[:])

	if _, err := rs.ReadAt(doc.no[:], offset+4); err != nil {
		return err
	}
	return nil
}
