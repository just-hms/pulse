package readers

import (
	"bufio"
	"io"
	"log"
	"strings"

	"github.com/just-hms/pulse/pkg/entity"
	"github.com/just-hms/pulse/pkg/index"
)

var _ index.ChunkReader = &msMarco{}

type msMarco struct {
	chunkSize    int
	reader       *bufio.Reader
	indexCounter uint
}

func NewMsMarco(r *bufio.Reader, c int) *msMarco {
	return &msMarco{
		chunkSize:    c,
		reader:       r,
		indexCounter: 0,
	}
}

// Read implements index.ChunkReader.
func (r *msMarco) Read() ([]*entity.Document, error) {
	res := make([]*entity.Document, r.chunkSize)

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

		res[i] = &entity.Document{
			ID:      r.indexCounter,
			No:      sPlitted[0],
			Content: strings.TrimSpace(sPlitted[1]),
		}
		r.indexCounter++
	}

	return res, nil
}
