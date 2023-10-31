package index

import (
	"runtime"
	"sync"

	"github.com/armon/go-radix"
	"github.com/just-hms/pulse/pkg/entity"
	"github.com/just-hms/pulse/pkg/word"
)

func Load(r ChunkReader) {
	numWorkers := runtime.GOMAXPROCS(-1)
	// Buffered channel
	chunksQueue := make(chan []*entity.Document, numWorkers)

	var wg sync.WaitGroup

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		i := i
		go func(workerID int) {
			defer wg.Done()
			for chunk := range chunksQueue {
				lessiconIndex := radix.New()

				for _, doc := range chunk {
					tokens := word.Tokenize(doc.Content)
					// TODO: add setting for this
					word.StopWordsRemoval(tokens)
					word.Stem(tokens)
					for _, t := range tokens {
						lessiconIndex.Insert(t, []int{})
					}
				}

				kek := map[string][]string{}
				// serialize the tree in a file
				lessiconIndex.WalkPath("", func(s string, v interface{}) bool {
					val := v.([]string)
					kek[s] = val
					return false
				})

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
