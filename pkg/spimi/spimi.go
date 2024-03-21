package spimi

import (
	"runtime"
	"sync"
	"time"

	"github.com/just-hms/pulse/pkg/spimi/inverseindex"
	"github.com/just-hms/pulse/pkg/spimi/preprocess"
	"github.com/just-hms/pulse/pkg/spimi/spimireader"
)

func Parse(r spimireader.Chunk, numWorkers int) error {
	chunksQueue := make(chan []spimireader.Document, numWorkers)
	consumerAvailable := make(chan bool, numWorkers)

	errChan := make(chan error, 1)

	b := NewBuilder()

	var wg sync.WaitGroup
	wg.Add(numWorkers)
	for i := 0; i < numWorkers; i++ {
		go func() {
			defer wg.Done()

			consumerAvailable <- true

			for chunk := range chunksQueue {

				for _, doc := range chunk {
					// for every chunk

					tokens := preprocess.GetTokens(doc.Content)
					freqs := make(map[string]uint32, 100)
					for _, term := range tokens {
						if _, ok := freqs[term]; !ok {
							freqs[term] = 1
						} else {
							freqs[term]++
						}
					}

					b.Lock()
					b.Add(freqs, inverseindex.Document{
						No:  doc.No,
						Len: len(doc.Content),
					})
					b.Unlock()
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
				errChan <- err
				return
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
		case err := <-errChan:
			return err
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

		b.Lock()
		err := b.Encode("data/dump")
		if err != nil {
			return err
		}
		b.Unlock()

		if !ok {
			break
		}
	}

	return nil
}
