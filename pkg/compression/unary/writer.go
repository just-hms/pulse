package unary

import (
	"encoding/binary"
	"io"
)

// Writer encodes data into unary format and writes as bits
type Writer struct {
	writer io.Writer
	buffer byte  // Accumulate bits before writing
	count  uint8 // Number of bits currently in buffer
}

// Write encodes the input as unary compression and writes to the underlying writer
func (ubw *Writer) Write(data []byte) (int, error) {
	// todo: should directly receive uint64
	result := binary.LittleEndian.Uint64(data)

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
		ubw.buffer |= 1 << (7 - ubw.count) // Set the corresponding bit
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

// Flush writes any remaining bits in the buffer and resets the state
func (ubw *Writer) Flush() error {
	if ubw.count > 0 {
		// Write the buffer to the underlying writer
		if _, err := ubw.writer.Write([]byte{ubw.buffer}); err != nil {
			return err
		}
		// Reset the buffer and bit count
		ubw.buffer = 0
		ubw.count = 0
	}
	return nil
}
