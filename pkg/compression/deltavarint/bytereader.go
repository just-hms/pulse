package deltavarint

import "io"

// byteReader turns an io.Reader into an io.ByteReader.
type byteReader struct {
	io.Reader
}

// newByteReader returns an io.ByteReader from an io.Reader
func newByteReader(r io.Reader) io.ByteReader {
	if r, ok := r.(io.ByteReader); ok {
		return r
	}
	return &byteReader{r}
}

func (r *byteReader) ReadByte() (byte, error) {
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
