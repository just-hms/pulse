package bytex

import "io"

// Reader turns an io.Reader into an io.ByteReader.
type Reader struct {
	io.Reader
}

// NewReader returns an io.ByteReader from an io.byteReader.
// If r is already a ByteReader, it is returned.
func NewReader(r io.Reader) io.ByteReader {
	if r, ok := r.(io.ByteReader); ok {
		return r
	}
	return &Reader{r}
}

func (r *Reader) ReadByte() (byte, error) {
	var buf [1]byte
	n, err := r.Reader.Read(buf[:])
	if err != nil {
		return 0, err
	}
	if n == 0 {
		return 0, io.ErrNoProgress
	}
	return buf[0], nil
}
