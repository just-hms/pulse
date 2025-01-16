package unary

import (
	"errors"
	"io"
)

type UnaryReader struct {
	reader io.Reader
	buffer byte  // Buffer to hold the current byte
	count  uint8 // Current bit position in the buffer
}

// Read decodes unary data and writes the decoded integers to the output slice `p`.
// It returns the number of integers written and any errors encountered.
func (ubr *UnaryReader) Read(p []byte) (int, error) {
	var values []byte

	for len(values) < len(p) {
		value, err := ubr.readUnary()
		if err != nil {
			if errors.Is(err, io.EOF) && len(values) > 0 {
				break // Return successfully if we already have decoded values
			}
			return len(values), err
		}
		values = append(values, byte(value))
	}

	// Copy decoded values to the output slice
	copy(p, values)
	return len(values), nil
}

// readUnary decodes a single unary-encoded value from the bitstream
func (ubr *UnaryReader) readUnary() (uint64, error) {
	var count uint64

	for {
		bit, err := ubr.readBit()
		if err != nil {
			return 0, err
		}
		if bit == 0 {
			// Terminating bit for unary encoding
			break
		}
		count++
	}

	return count, nil
}

// readBit reads a single bit from the buffer
func (ubr *UnaryReader) readBit() (int, error) {
	if ubr.count == 0 {
		// Refill the buffer if all bits are consumed
		buf := make([]byte, 1)
		_, err := ubr.reader.Read(buf)
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
