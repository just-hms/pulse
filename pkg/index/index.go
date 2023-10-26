package index

import (
	"bufio"
	"fmt"
	"io"

	"github.com/just-hms/pulse/pkg/entity"
)

func Load(reader *bufio.Reader, withs ...With) {
	defaults := &Option{
		ChunkSize: 100,
	}
	for _, w := range withs {
		w(defaults)
	}

	for {
		chunk, err := readChunk(reader, defaults.ChunkSize)
		if err != nil {
			panic(err)
		}
		fmt.Println(chunk)
	}
}

// TODO: get the readChunk as dependency
func readChunk(reader *bufio.Reader, chunkSize int) ([]*entity.Document, error) {
	res := make([]*entity.Document, chunkSize)
	for i := 0; i < chunkSize; i++ {
		var err error
		res[i].Content, err = reader.ReadString('\n')
		if err == io.EOF {
			return res, nil
		}
		if err != nil {
			return nil, err
		}
	}
	return res, nil
}
