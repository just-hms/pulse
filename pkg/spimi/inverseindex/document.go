package inverseindex

import (
	"encoding/binary"
	"io"
	"unsafe"
)

type Document struct {
	No   [8]byte
	Size uint32
}

const DOC_SIZE = unsafe.Sizeof(Document{})

func NewDocument(no string, size int) Document {
	// todo: check this (strange)
	if len(no) > 8 {
		panic("dockNo cannot be bigger then 8")
	}
	d := Document{
		Size: uint32(size),
	}
	copy(d.No[:], no)
	return d
}

func (doc *Document) Encode(w io.Writer) error {
	if err := binary.Write(w, binary.LittleEndian, doc.Size); err != nil {
		return err
	}

	if err := binary.Write(w, binary.LittleEndian, doc.No); err != nil {
		return err
	}
	return nil
}

func (doc *Document) Decode(ID uint32, rs io.ReaderAt) error {
	offset := int64(DOC_SIZE) * int64(ID)

	var buf [4]byte

	if _, err := rs.ReadAt(buf[:], offset+0); err != nil {
		return err
	}
	doc.Size = binary.LittleEndian.Uint32(buf[:])

	if _, err := rs.ReadAt(doc.No[:], offset+4); err != nil {
		return err
	}
	return nil
}
