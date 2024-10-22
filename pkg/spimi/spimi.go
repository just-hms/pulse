package spimi

import (
	"os"
	"path/filepath"
	"runtime"
	"time"

	iradix "github.com/hashicorp/go-immutable-radix/v2"
	"github.com/just-hms/pulse/pkg/preprocess"
	"github.com/just-hms/pulse/pkg/spimi/inverseindex"
	"github.com/just-hms/pulse/pkg/structures/radix"
	"golang.org/x/sync/errgroup"
)

const (
	sleep                  = 10 * time.Second
	memoryThreshold uint64 = 4 * 1024 * 1024 * 1024
)

func Parse(r ChunkReader, numWorkers int, path string) error {

	var (
		workers  errgroup.Group
		dumper   errgroup.Group
		b        = newBuilder()
		memStats runtime.MemStats
	)

	workers.SetLimit(numWorkers)

	dumper.Go(func() error {
		for {
			time.Sleep(sleep)

			runtime.ReadMemStats(&memStats)
			eof := r.EOF()

			if memStats.Alloc < memoryThreshold && !eof {
				continue
			}

			if eof {
				err := workers.Wait()
				if err != nil {
					return err
				}
			}

			err := b.Encode(path)
			if err != nil {
				return err
			}
			if eof {
				return nil
			}
		}
	})

	for chunk, err := range r.Read() {
		if err != nil {
			return err
		}
		workers.Go(func() error {
			for _, doc := range chunk {
				tokens := preprocess.GetTokens(doc.Content)
				freqs := make(map[string]uint32, len(tokens)/2)
				for _, term := range tokens {
					if _, ok := freqs[term]; !ok {
						freqs[term] = 1
					} else {
						freqs[term]++
					}
				}

				b.Add(freqs, inverseindex.NewDocument(doc.No, len(doc.Content)))
			}
			return nil
		})
	}

	return dumper.Wait()
}

func Merge(path string) error {
	partitions, err := os.ReadDir(path)
	if err != nil {
		return err
	}

	gLexicon := iradix.New[*inverseindex.GlobalTerm]().Txn()

	for _, partition := range partitions {
		if !partition.IsDir() {
			continue
		}
		folder := filepath.Join(path, partition.Name())
		f, err := os.Open(filepath.Join(folder, "terms.bin"))
		if err != nil {
			return err
		}

		lLexicon := iradix.New[*inverseindex.GlobalTerm]()
		err = radix.Decode(f, &lLexicon)
		if err != nil {
			return err
		}

		// append values to gLexicon
		for lk, lv := range radix.Values(lLexicon) {
			if gv, ok := gLexicon.Get(lk); ok {
				gv.Frequence += lv.Frequence
				gv.Appearences += lv.Appearences
			} else {
				gLexicon.Insert(lk, lv)
			}
		}

	}

	f, err := os.Create(filepath.Join(path, "terms.bin"))
	if err != nil {
		return err
	}
	defer f.Close()

	return radix.Encode(f, gLexicon.Commit())
}
