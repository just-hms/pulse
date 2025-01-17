package spimi

import (
	"encoding/gob"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"time"

	iradix "github.com/hashicorp/go-immutable-radix/v2"
	"github.com/just-hms/pulse/pkg/countwriter"
	"github.com/just-hms/pulse/pkg/preprocess"
	"github.com/just-hms/pulse/pkg/spimi/inverseindex"
	"github.com/just-hms/pulse/pkg/structures/radix"
	"github.com/just-hms/pulse/pkg/structures/withkey"
	"golang.org/x/exp/mmap"
	"golang.org/x/sync/errgroup"
)

const (
	KB = 1024
	MB = KB * 1024
	GB = MB * 1024
)

const DefaulMemoryThreshold = 3 * GB / MB

func (b *builder) Parse(r ChunkReader, numWorkers int, path string) error {

	var (
		workers  errgroup.Group
		dumper   errgroup.Group
		memStats runtime.MemStats
	)

	workers.SetLimit(numWorkers)

	dumper.Go(func() error {
		hit := 0
		for {
			time.Sleep(time.Second)

			runtime.ReadMemStats(&memStats)
			eof := r.EOF()

			if memStats.Alloc > uint64(b.IndexingSettings.MemoryThresholdMB*MB) {
				hit++
			} else {
				hit = 0
			}

			if hit < 3 && !eof {
				continue
			}

			hit = 0

			if eof {
				err := workers.Wait()
				if err != nil {
					return err
				}
			}

			err := b.encode(path)
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
				tokens := preprocess.GetTokens(doc.Content, b.IndexingSettings)
				freqs := make(map[string]uint32, len(tokens)/2)
				for _, term := range tokens {
					if _, ok := freqs[term]; !ok {
						freqs[term] = 1
					} else {
						freqs[term]++
					}
				}

				b.add(freqs, inverseindex.NewDocument(doc.No, len(doc.Content)))
			}
			return nil
		})
	}

	return dumper.Wait()
}

func Merge(path string) error {
	log.Println("merging...")
	defer log.Println("...end")
	partitions, err := ReadPartitions(path)
	if err != nil {
		return err
	}

	if len(partitions) == 0 {
		return nil
	}

	gTermInfoF, err := os.Create(filepath.Join(path, "terms-info.bin"))
	if err != nil {
		return err
	}
	defer gTermInfoF.Close()

	gTermF, err := os.Create(filepath.Join(path, "terms.bin"))
	if err != nil {
		return err
	}
	defer gTermF.Close()

	termsEnc := gob.NewEncoder(gTermF)

	gLexicon := iradix.New[uint32]().Txn()
	cw := countwriter.NewWriter(gTermInfoF)

	for _, partition := range partitions {
		folder := filepath.Join(path, partition.Name())

		termF, err := os.Open(filepath.Join(folder, "terms.bin"))
		if err != nil {
			return err
		}

		lTermInfoR, err := mmap.Open(filepath.Join(folder, "terms-info.bin"))
		if err != nil {
			return err
		}

		lLexicon := iradix.New[uint32]()

		if err := radix.Decode(termF, &lLexicon); err != nil {
			return err
		}

		// append values to gLexicon
		for key, lOffset := range radix.Values(lLexicon) {

			lTerm, err := inverseindex.DecodeTerm[inverseindex.LocalTerm](lOffset, lTermInfoR)
			if err != nil {
				return err
			}

			gOffset, ok := gLexicon.Get(key)
			if ok {
				// get the global value
				gTerm, err := inverseindex.DecodeTerm[inverseindex.GlobalTerm](gOffset, gTermInfoF)
				if err != nil {
					return err
				}

				gTerm.DocumentFrequency += lTerm.DocumentFrequency
				gTerm.MaxTermFrequency = max(gTerm.MaxTermFrequency, lTerm.MaxTermFrequency)

				if err := inverseindex.UpdateTerm(gTerm, gOffset, gTermInfoF); err != nil {
					return err
				}
			} else {
				gTerm := &withkey.WithKey[inverseindex.GlobalTerm]{
					Key: string(key),
					Value: inverseindex.GlobalTerm{
						DocumentFrequency: lTerm.DocumentFrequency,
						MaxTermFrequency:  lTerm.MaxTermFrequency,
					},
				}

				gLexicon.Insert(key, uint32(cw.Count))

				if err := inverseindex.EncodeTerm(gTerm, termsEnc, cw); err != nil {
					return err
				}
			}
		}

	}

	return nil
}
