package unary

import (
	"encoding/binary"
	"io"
)

// Writer encodes data into unary format and writes as bits
type Writer struct {
	io.Writer

	padded [8]byte

	buffer byte  // Accumulate bits before writing
	count  uint8 // Number of bits currently in buffer
}

func NewWriter(w io.Writer) *Writer {
	return &Writer{
		Writer: w,
		padded: [8]byte{},
	}
}

// Write encodes the input as unary compression and writes to the underlying writer
func (ubw *Writer) Write(data []byte) (int, error) {
	copy(ubw.padded[:len(data)], data)

	result := binary.LittleEndian.Uint64(ubw.padded[:])

	for i := uint64(0); i < result; i++ {
		if err := ubw.writeBit(1); err != nil {
			return 0, err
		}
	}

	// Write terminating `0`
	if err := ubw.writeBit(0); err != nil {
		return 0, err
	}

	return len(data), nil
}

// writeBit writes a single bit to the buffer
func (ubw *Writer) writeBit(bit int) error {
	if bit != 0 {
		ubw.set()
	}
	ubw.count++

	// If the buffer is full (8 bits), write it to the writer
	if ubw.count < 8 {
		return nil
	}

	if err := ubw.Flush(); err != nil {
		return err
	}

	return nil
}

func (ubw *Writer) set() {
	ubw.buffer |= 1 << (7 - ubw.count) // Set the corresponding bit
}

// Flush writes any remaining bits in the buffer and resets the state
func (ubw *Writer) Flush() error {
	if ubw.count > 0 {

		// use 11111 as padding so if the reader find no zero it returns with EOF
		for ; ubw.count <= 8; ubw.count++ {
			ubw.set()
		}

		// Write the buffer to the underlying writer
		if _, err := ubw.Writer.Write([]byte{ubw.buffer}); err != nil {
			return err
		}
		// Reset the buffer and bit count
		ubw.buffer = 0
		ubw.count = 0
	}
	return nil
}
