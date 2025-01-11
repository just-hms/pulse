package readers

import (
	"bufio"
	"errors"
	"io"
	"iter"
	"log"
	"strings"

	"github.com/just-hms/pulse/pkg/spimi"
)

var _ spimi.ChunkReader = &msMarco{}

type msMarco struct {
	chunkSize int
	reader    *bufio.Reader
	eof       bool
}

func NewMsMarco(r *bufio.Reader, chunkSize int) *msMarco {
	return &msMarco{
		reader:    r,
		chunkSize: chunkSize,
	}
}

func (r *msMarco) EOF() bool {
	return r.eof
}

// Read implements index.ChunkReader.
func (r *msMarco) Read() iter.Seq2[[]spimi.Document, error] {

	return func(yield func([]spimi.Document, error) bool) {

		for {
			res := make([]spimi.Document, 0, r.chunkSize)
			for i := 0; i < r.chunkSize; i++ {
				s, err := r.reader.ReadString('\n')
				if errors.Is(err, io.EOF) {
					r.eof = true
					yield(res, nil)
					return
				}
				if err != nil {
					yield(nil, err)
					return
				}

				sPlitted := strings.Split(s, "\t")
				if len(sPlitted) != 2 {
					log.Fatalf("'%s' is malformed\n", strings.TrimSpace(s))
				}
				res = append(res, spimi.Document{
					No:      sPlitted[0],
					Content: strings.TrimSpace(sPlitted[1]),
				})
			}

			if !yield(res, nil) {
				return
			}
		}
	}
}
