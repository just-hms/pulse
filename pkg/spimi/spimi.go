package spimi

import (
	"encoding/gob"
	"errors"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"sync"
	"time"

	"github.com/just-hms/pulse/pkg/preprocess"
	"github.com/just-hms/pulse/pkg/spimi/inverseindex"
)

func Parse(r ChunkReader, numWorkers int, path string) error {
	chunksQueue := make(chan []Document, numWorkers)
	consumerAvailable := make(chan bool, numWorkers)

	errChan := make(chan error, 1)

	b := newBuilder()

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
					freqs := make(map[string]uint32, len(tokens)/2)
					for _, term := range tokens {
						if _, ok := freqs[term]; !ok {
							freqs[term] = 1
						} else {
							freqs[term]++
						}
					}

					b.Add(freqs, inverseindex.Document{
						No:  doc.No,
						Len: len(doc.Content),
					})
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
	const memoryThreshold uint64 = 8 * 1024 * 1024 * 1024

	sleep := 10 * time.Second
	for {
		var working = true
		select {
		case err := <-errChan:
			return err
		case _, working = <-chunksQueue:
			if working {
				time.Sleep(sleep)
			} else {
				wg.Wait()
			}
		default:
			time.Sleep(sleep)
		}

		runtime.ReadMemStats(&memStats)
		if memStats.Alloc < memoryThreshold && working {
			continue
		}

		err := b.Encode(path)
		if err != nil {
			return err
		}

		if !working {
			break
		}
	}

	return nil
}

func Merge(path string) error {
	partitions, err := os.ReadDir(path)
	if err != nil {
		return err
	}

	// TODO: maybe do this without keeping everything in memory
	// TODO: add a "global term" which doesn't have local indexes information
	res := []inverseindex.Term{}
	for _, partition := range partitions {
		if !partition.IsDir() {
			continue
		}
		folder := filepath.Join(path, partition.Name())
		f, err := os.Open(filepath.Join(folder, "terms.bin"))
		if err != nil {
			return err
		}

		reader := gob.NewDecoder(f)

		for {
			t := inverseindex.Term{}
			err = reader.Decode(&t)
			if errors.Is(err, io.EOF) {
				break
			}
			if err != nil {
				return err
			}
			res = append(res, t)
		}
	}
	slices.SortFunc(res, func(a, b inverseindex.Term) int {
		if a.Value == b.Value {
			return 0
		}
		if a.Value < b.Value {
			return -1
		}
		return 1
	})

	f, err := os.Create(filepath.Join(path, "terms.bin"))
	if err != nil {
		return err
	}
	encoder := gob.NewEncoder(f)

	for _, t := range res {
		err := encoder.Encode(t)
		if err != nil {
			return err
		}
	}

	return nil
}
