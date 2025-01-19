package unary

import (
	"encoding/binary"
	"io"
)

// Reader given a io.Reader and a minimum reads out a uint number given its unary representation
type Reader struct {
	io.Reader
	minimum uint

	padded [8]byte

	buffer byte  // Buffer to hold the current byte
	count  uint8 // Current bit position in the buffer
}

func NewReader(r io.Reader, m uint) *Reader {
	return &Reader{
		Reader:  r,
		padded:  [8]byte{},
		minimum: m,
	}
}

// Read decodes unary data and writes the decoded integers to the output slice `p`
func (ubr *Reader) Read(p []byte) (int, error) {
	value, err := ubr.readUnary()
	if err != nil {
		return 0, err
	}

	binary.LittleEndian.PutUint64(ubr.padded[:], value)
	copy(p, ubr.padded[:])

	return len(p), nil
}

// readUnary decodes a single unary-encoded value from the bitstream
func (ubr *Reader) readUnary() (uint64, error) {
	var count uint64

	for {
		bit, err := ubr.readBit()
		if err != nil {
			return 0, err
		}
		if bit == 0 {
			break
		}
		count++
	}

	return count + uint64(ubr.minimum), nil
}

// readBit reads a single bit from the buffer
func (ubr *Reader) readBit() (int, error) {
	if ubr.count == 0 {
		// Refill the buffer if all bits are consumed
		buf := make([]byte, 1)
		_, err := ubr.Reader.Read(buf)
		if err != nil {
			return 0, err
		}
		ubr.buffer = buf[0]
		ubr.count = 8
	}

	// Extract the current bit
	bit := (ubr.buffer >> 7) & 1
	ubr.buffer <<= 1
	ubr.count--

	return int(bit), nil
}
