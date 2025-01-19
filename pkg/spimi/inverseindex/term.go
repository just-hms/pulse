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

// GlobalTerm contians the global information about a term
type GlobalTerm struct {
	DocumentFrequency uint32 // number of documents in which the term appears
	MaxTermFrequency  uint32 // max frequence inside a document
}

// LocalTerm contains the information about a term in a given document partition
type LocalTerm struct {
	GlobalTerm

	FreqStart, FreqLength uint32

	PostStart, PostLength uint32
}

// DiscRef is a pairt TermKey -> Pointer to a location in disk where the information of a term are stored
type DiscRef withkey.WithKey[uint32]

// EncodeTerm encodes a term using a gob encoder and stores its information into a countwriter.Writer
func EncodeTerm[T any](t *withkey.WithKey[T], termW *gob.Encoder, cw *countwriter.Writer) error {
	// encode the term and set its pointer to the last written byte position
	if err := termW.Encode(DiscRef{t.Key, uint32(cw.Count)}); err != nil {
		return err
	}

	// encode the term info
	if err := binary.Write(cw, binary.LittleEndian, t.Value); err != nil {
		return err
	}

	return nil
}

func UpdateTerm[T any](t T, offset uint32, w io.WriterAt) error {
	// determine the size of T
	size := int(unsafe.Sizeof(t))

	var buf bytes.Buffer

	// write t inside a buffer
	if err := binary.Write(&buf, binary.LittleEndian, t); err != nil {
		return err
	}
	// cap the buffer to the size
	data := buf.Bytes()[:size]

	// write down the buffer in the given position
	if _, err := w.WriteAt(data, int64(offset)); err != nil {
		return err
	}
	return nil
}

// DecodeTerm reads and decodes a term of type T from the provided reader at the offset specified by t.
func DecodeTerm[T any](offset uint32, r io.ReaderAt) (*T, error) {
	var result T

	// determine the size of the type T
	size := int(unsafe.Sizeof(result))

	// create a byte slice to hold the data
	data := make([]byte, size)

	// read data from the specified offset
	if _, err := r.ReadAt(data, int64(offset)); err != nil {
		return nil, err
	}

	// encode the data into the result element
	if err := binary.Read(bytes.NewReader(data), binary.LittleEndian, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
