package countwriter

import (
	"io"
)

type Writer struct {
	io.Writer
	Count int
}

func NewWriter(w io.Writer) *Writer {
	return &Writer{
		Writer: w,
		Count:  0,
	}
}

func (cw *Writer) Write(p []byte) (n int, err error) {
	n, err = cw.Writer.Write(p)
	cw.Count += n
	return n, err
}
