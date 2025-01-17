package countwriter

import (
	"io"
	"os"
)

type Writer struct {
	*os.File
	Count int
}

func NewWriter(f *os.File) *Writer {
	return &Writer{
		File:  f,
		Count: 0,
	}
}

func (cw *Writer) Write(p []byte) (n int, err error) {
	// todo: check if this is necessary
	if _, err := cw.File.Seek(0, io.SeekEnd); err != nil {
		return 0, err
	}
	n, err = cw.File.Write(p)
	cw.Count += n
	return n, err
}
