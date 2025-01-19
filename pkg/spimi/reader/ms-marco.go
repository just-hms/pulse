package reader

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

// msMarco reader responsible to read msMarco passenger dataset
type msMarco struct {
	chunkSize int
	reader    *bufio.Reader
	eof       bool
}

// NewMsMarco returns an msMarco reader
func NewMsMarco(r *bufio.Reader, chunkSize int) *msMarco {
	return &msMarco{
		reader:    r,
		chunkSize: chunkSize,
	}
}

// EOF returns true if the reader has finished reading
func (r *msMarco) EOF() bool {
	return r.eof
}

// Read implements index.ChunkReader and returns an iterator containing the chunks from a dataset
func (r *msMarco) Read() iter.Seq2[spimi.Chunk, error] {

	return func(yield func(spimi.Chunk, error) bool) {

		for {
			// create a chunk with r.chunkSize capacity and 0 length
			res := make(spimi.Chunk, 0, r.chunkSize)

			// read line by line
			for i := 0; i < r.chunkSize; i++ {
				s, err := r.reader.ReadString('\n')
				// if the file is ended return the chunk and set r.eof to true
				if errors.Is(err, io.EOF) {
					r.eof = true
					yield(res, nil)
					return
				}
				// if and error raised return it
				if err != nil {
					yield(nil, err)
					return
				}

				// spit the line by \t
				sPlitted := strings.Split(s, "\t")

				if len(sPlitted) != 2 {
					log.Fatalf("'%s' is malformed\n", strings.TrimSpace(s))
				}

				// append a document to the chunk
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
