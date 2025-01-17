package inverseindex

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"io"
	"unsafe"

	"github.com/just-hms/pulse/pkg/countwriter"
	"github.com/just-hms/pulse/pkg/structures/withkey"
)

type GlobalTerm struct {
	DocumentFrequency uint32 // number of documents in which the term appears
	MaxTermFrequency  uint32 // max frequence inside a document
}

// todo: data of GlobalTerm inside LocalTerm should not be used
type LocalTerm struct {
	GlobalTerm

	FreqStart, FreqLength uint32

	PostStart, PostLength uint32
}

type DiscRef withkey.WithKey[uint32]

func EncodeTerm[T any](t *withkey.WithKey[T], termW *gob.Encoder, cw *countwriter.Writer) error {
	if err := termW.Encode(DiscRef{t.Key, uint32(cw.Count)}); err != nil {
		return err
	}

	if err := binary.Write(cw, binary.LittleEndian, t.Value); err != nil {
		return err
	}

	return nil
}

func UpdateTerm[T any](t T, offset uint32, w io.WriterAt) error {
	// Determine the size of the type T
	size := int(unsafe.Sizeof(t))
	data := make([]byte, size)

	if _, err := w.WriteAt(data, int64(offset)); err != nil {
		return err
	}
	return nil
}

// DecodeTerm reads and decodes a term of type T from the provided reader at the offset specified by t.
func DecodeTerm[T any](offset uint32, r io.ReaderAt) (*T, error) {
	var result T

	// Determine the size of the type T
	size := int(unsafe.Sizeof(result))

	// Create a byte slice to hold the data
	data := make([]byte, size)

	// Read data from the offset specified in DiscRef
	if _, err := r.ReadAt(data, int64(offset)); err != nil {
		return nil, err
	}

	if err := binary.Read(bytes.NewReader(data), binary.LittleEndian, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
