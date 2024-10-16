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

func (doc *Document) Decode(ID uint32, rs io.ReadSeeker) error {
	if _, err := rs.Seek(int64(DOC_SIZE)*int64(ID), 0); err != nil {
		return err
	}
	if err := binary.Read(rs, binary.LittleEndian, &doc.Size); err != nil {
		return err
	}
	if err := binary.Read(rs, binary.LittleEndian, &doc.No); err != nil {
		return err
	}
	return nil
}
