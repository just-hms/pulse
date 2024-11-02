package bufseekio

import (
	"io"
)

var _ io.ReadSeeker = &reader{}

type reader struct {
	src     io.ReadSeeker
	buf     []byte
	bufSize int

	bufStart int64
	srcPos   int64

	srcEndPos *int64
	filled    bool
}

// NewReaderSize initializes a bufSeeker with an io.ReadSeeker and a buffer size.
func NewReaderSize(src io.ReadSeeker, bufSize int) *reader {
	return &reader{
		src:     src,
		bufSize: bufSize,
		buf:     make([]byte, bufSize),
	}
}

// NewReaderSize initializes a bufSeeker with an io.ReadSeeker and a buffer size.
func NewReaderBuffer(src io.ReadSeeker, buf []byte) *reader {
	return &reader{
		src:     src,
		bufSize: len(buf),
		buf:     buf,
	}
}

// NewReader initializes a bufSeeker with an io.ReadSeeker.
func NewReader(src io.ReadSeeker) *reader {
	return &reader{
		src:     src,
		bufSize: 4096,
		buf:     make([]byte, 4096),
	}
}

func (b *reader) Seek(offset int64, whence int) (int64, error) {
	switch whence {
	case io.SeekStart:
		b.srcPos = offset
	case io.SeekCurrent:
		b.srcPos = b.srcPos + offset
	case io.SeekEnd:
		if b.srcEndPos == nil {
			var err error
			srcEnd, err := b.src.Seek(0, io.SeekEnd)
			if err != nil {
				return 0, err
			}
			b.srcEndPos = &srcEnd
		}
		b.srcPos = *b.srcEndPos + offset
	default:
		panic("invalid whence")
	}

	return b.srcPos, nil
}

func (b *reader) Read(p []byte) (int, error) {
	var (
		n   = 0
		err error
	)

	for {
		bufPos := b.srcPos - b.bufStart

		if bufPos < 0 || bufPos >= int64(len(b.buf)) || !b.filled {
			err = b.refillBuffer()
			if err != nil && err != io.EOF {
				return 0, err
			}
			bufPos = 0
		}

		partialN := copy(p[n:], b.buf[bufPos:])
		b.srcPos += int64(partialN)
		n += partialN

		if n == len(p) {
			return n, nil
		}
		if err == io.EOF {
			return n, err
		}
	}
}

// refillBuffer refills the buffer starting at the current source position.
func (b *reader) refillBuffer() error {
	b.filled = true

	_, err := b.src.Seek(b.srcPos, io.SeekStart)
	if err != nil {
		return err
	}

	n, err := b.src.Read(b.buf)
	if err != nil && err != io.EOF {
		return err
	}

	b.buf = b.buf[:n]
	b.bufStart = b.srcPos
	return err
}
