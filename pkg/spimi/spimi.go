package spimi

import (
	"runtime"
	"sync"
	"time"

	"github.com/just-hms/pulse/pkg/spimi/info"
	"github.com/just-hms/pulse/pkg/word"
)

func Load(r ChunkReader, numWorkers int) []string {
	chunksQueue := make(chan []Document, numWorkers)
	consumerAvailable := make(chan bool, numWorkers)

	inf := info.NewSpimiBuilder()

	var wg sync.WaitGroup
	wg.Add(numWorkers)
	for i := 0; i < numWorkers; i++ {
		go func() {
			defer wg.Done()

			consumerAvailable <- true

			for chunk := range chunksQueue {

				for _, doc := range chunk {
					// for every chunk
					tokens := word.Tokenize(doc.Content)
					tokens = word.StopWordsRemoval(tokens)
					word.Stem(tokens)

					freqs := make(map[string]uint32, 100)
					for _, term := range tokens {
						if _, ok := freqs[term]; !ok {
							freqs[term] = 1
						} else {
							freqs[term]++
						}
					}

					inf.Lock()
					inf.Add(freqs, info.Doc{
						No:  doc.No,
						Len: len(doc.Content),
					})
					inf.Unlock()
				}
				consumerAvailable <- true
			}
		}()
	}

	// producer
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
			<-consumerAvailable
			chunksQueue <- chunk
		}
	}()

	// dumper
	var memStats runtime.MemStats
	var memoryThreshold uint64 = 8 * 1024 * 1024 * 1024

	sleep := 10 * time.Second
	for {
		var ok = true
		select {
		case _, ok = <-chunksQueue:
			if ok {
				time.Sleep(sleep)
			} else {
				wg.Wait()
			}
		default:
			time.Sleep(sleep)
		}

		runtime.ReadMemStats(&memStats)
		if memStats.Alloc < memoryThreshold && ok {
			continue
		}

		inf.Lock()
		err := inf.Dump("data/dump")
		if err != nil {
			panic(err)
		}
		inf.Unlock()

		if !ok {
			break
		}
	}

	// Wait for all workers to finish

	return []string{}
}
