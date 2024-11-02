package bufseekio

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReadWithinBuffer(t *testing.T) {
	t.Parallel()
	req := require.New(t)

	src := strings.NewReader("Hello, bufseekio!")
	reader := NewReaderSize(src, 8)

	buf := make([]byte, 5)
	n, err := reader.Read(buf)
	req.NoError(err)
	req.Equal("Hello", string(buf[:n]))
}

func TestReadAcrossBufferBoundary(t *testing.T) {
	t.Parallel()
	req := require.New(t)

	src := strings.NewReader("Hello, bufseekio!")
	reader := NewReaderSize(src, 8)

	buf := make([]byte, 15)
	n, err := reader.Read(buf)

	req.NoError(err)
	req.Equal("Hello, bufseeki", string(buf[:n]))
}

func TestSeekWithinBuffer(t *testing.T) {
	t.Parallel()
	req := require.New(t)

	src := strings.NewReader("Hello, bufseekio!")
	reader := NewReaderSize(src, 8)

	_, _ = reader.Read(make([]byte, 5)) // Read some data to fill the buffer

	pos, err := reader.Seek(2, io.SeekStart)
	req.NoError(err)
	req.Equal(int64(2), pos)

	buf := make([]byte, 3)
	n, err := reader.Read(buf)
	req.NoError(err)

	req.Equal("llo", string(buf[:n]))
}

func TestSeekOutsideBuffer(t *testing.T) {
	t.Parallel()
	req := require.New(t)

	src := strings.NewReader("Hello, bufseekio!")
	reader := NewReaderSize(src, 8)

	// Seek outside the initial buffer range
	pos, err := reader.Seek(10, io.SeekStart)
	req.NoError(err)
	req.Equal(int64(10), pos)

	buf := make([]byte, 5)
	n, err := reader.Read(buf)
	req.NoError(err)

	req.NoError(err)
	req.Equal("seeki", string(buf[:n]))
}

func TestSeekToEnd(t *testing.T) {
	t.Parallel()
	req := require.New(t)

	base := "Hello, bufseekio!"
	src := strings.NewReader(base)
	reader := NewReaderSize(src, 8)

	// Seek to the end
	pos, err := reader.Seek(0, io.SeekEnd)
	req.NoError(err)
	req.Equal(pos, int64(len(base)))

	// Trying to read at the end should return EOF
	buf := make([]byte, 5)
	n, err := reader.Read(buf)
	req.Equal(io.EOF, err)
	req.Equal(0, n)
}

func TestReadToEOF(t *testing.T) {
	t.Parallel()
	req := require.New(t)

	base := "Hello, bufseekio!"

	src := strings.NewReader(base)
	reader := NewReaderSize(src, 8)

	var result bytes.Buffer
	buf := make([]byte, 5)

	for {
		n, err := reader.Read(buf)
		result.Write(buf[:n])
		if err == io.EOF {
			break
		}
		req.NoError(err)
	}

	req.Equal(base, result.String())
}
