package deltavarint

import (
	"encoding/binary"
	"io"
)

type Writer struct {
	io.Writer
	last   uint64
	padded [8]byte
}

// NewWriter creates a new delta-varint writer wrapping the provided ByteWriter.
func NewWriter(w io.Writer) *Writer {
	return &Writer{
		Writer: w,
		last:   0,
		padded: [8]byte{},
	}
}

// Write encodes the provided buffer of uint64 values as delta-encoded varints.
func (vbw *Writer) Write(data []byte) (int, error) {
	copy(vbw.padded[:len(data)], data)
	value := binary.LittleEndian.Uint64(vbw.padded[:])

	// Calculate the delta
	delta := value - vbw.last
	vbw.last = value

	// Encode the delta as a varint
	var buf [binary.MaxVarintLen64]byte
	toWrite := binary.PutUvarint(buf[:], delta)

	n, err := vbw.Writer.Write(buf[:toWrite])
	if err != nil {
		return 0, err
	}

	return n, nil
}
