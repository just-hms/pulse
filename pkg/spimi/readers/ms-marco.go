package readers

import (
	"bufio"
	"io"
	"log"
	"strings"

	"github.com/just-hms/pulse/pkg/spimi"
)

var _ spimi.ChunkReader = &msMarco{}

type msMarco struct {
	chunkSize int
	reader    *bufio.Reader
}

func NewMsMarco(r *bufio.Reader, chunkSize int) *msMarco {
	return &msMarco{
		reader:    r,
		chunkSize: chunkSize,
	}
}

// Read implements index.ChunkReader.
func (r *msMarco) Read() ([]spimi.Document, error) {
	res := make([]spimi.Document, r.chunkSize)

	for i := 0; i < r.chunkSize; i++ {
		s, err := r.reader.ReadString('\n')
		if err == io.EOF {
			return res[:i], nil
		}
		if err != nil {
			return nil, err
		}

		sPlitted := strings.Split(s, "\t")
		if len(sPlitted) != 2 {
			log.Fatalf("'%s' is malformed\n", strings.TrimSpace(s))
		}

		res[i] = spimi.Document{
			No:      sPlitted[0],
			Content: strings.TrimSpace(sPlitted[1]),
		}
	}
	return res, nil
}
