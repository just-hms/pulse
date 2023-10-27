package index

import (
	"encoding/json"
	"os"
	"runtime"
	"sync"

	"github.com/just-hms/pulse/pkg/entity"
	"github.com/just-hms/pulse/pkg/word"
)

type Index struct {
	file string
	// TODO: change this map in something like a trie
	// TODO: keep the posting list orderer
	// TODO: add the frequencies
	base map[string]*[]uint
}

func New() *Index {
	return &Index{
		file: "",
		// TODO: get me from the chunksize
		base: make(map[string]*[]uint, 100_000),
	}
}
func (i *Index) Add(doc *entity.Document, tokens []string) {
	for _, t := range tokens {
		// TODO: do less accesses
		if val, ok := i.base[t]; !ok {
			i.base[t] = &[]uint{doc.ID}
		} else {
			*val = append(*val, doc.ID)
		}
	}
}

func (i *Index) Flush() {
	var (
		f   *os.File
		err error
	)

	if i.file == "" {
		f, err = os.CreateTemp("", "")
		if err != nil {
			panic(err)
		}
	} else {
		f, err = os.Open(i.file)
		if err != nil {
			panic(err)
		}
	}
	res, err := json.Marshal(i.base)
	if err != nil {
		panic(err)
	}

	_, err = f.Write(res)
	if err != nil {
		panic(err)
	}
}

func Load(r ChunkReader) {
	numWorkers := runtime.GOMAXPROCS(-1)
	chunksQueue := make(chan []*entity.Document, numWorkers) // Buffered channel

	var wg sync.WaitGroup

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		i := i
		go func(workerID int) {
			defer wg.Done()
			for chunk := range chunksQueue {
				i := New()
				for _, doc := range chunk {
					tokens := word.Tokenize(doc.Content)
					// TODO: add setting for this
					word.StopWordsRemoval(tokens)
					word.Stem(tokens)
					i.Add(doc, tokens)
				}
				i.Flush()
			}
		}(i)
	}

	// readers
	go func() {
		defer close(chunksQueue)
		for {
			chunk, err := r.Read()
			if err != nil {
				panic(err)
			}
			if len(chunk) == 0 {
				return // Stop when there are no more chunks
			}
			chunksQueue <- chunk
		}
	}()

	// Wait for all workers to finish
	wg.Wait()
}
