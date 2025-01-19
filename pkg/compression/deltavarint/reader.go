package deltavarint

import (
	"encoding/binary"
	"io"
)

// given and io.Reader reads out uint number using variable bytes encoding of the delta between a list of numbers
type Reader struct {
	io.ByteReader
	last uint64

	padded [8]byte
}

func NewReader(r io.Reader) *Reader {
	return &Reader{
		ByteReader: newByteReader(r),
		last:       0,
		padded:     [8]byte{},
	}
}

// Read decodes the delta-encoded varints into the buffer, read bytes shouldn't be used
func (vbr *Reader) Read(p []byte) (int, error) {
	var (
		err   error
		value uint64
	)

	// Decode the next varint from the reader
	value, err = binary.ReadUvarint(vbr.ByteReader)
	if err != nil {
		return 0, err
	}

	// Apply the delta to the last value
	vbr.last += value

	// Write the decoded value to the buffer
	binary.LittleEndian.PutUint64(vbr.padded[:], vbr.last)
	copy(p, vbr.padded[:])
	return 0, nil
}
