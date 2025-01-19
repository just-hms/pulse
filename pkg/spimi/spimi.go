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

const (
	DefaulMemoryThreshold            = 3 * GB / MB
	MemoryThresholdHitsBeforeDumping = 3
)

// Parse parses a given dataset dumping the indexing information into the path
func (b *builder) Parse(r ChunkReader, path string) error {

	var (
		workers  errgroup.Group
		dumper   errgroup.Group
		memStats runtime.MemStats
	)

	workers.SetLimit(b.NumWorkers)

	// dumper goroutine responsible for dumping the indexing information
	dumper.Go(func() error {

		secondsAboveTheTresholdCounter := 0

		for {
			time.Sleep(time.Second)

			// read the memory info
			runtime.ReadMemStats(&memStats)

			// if the memory info exceeds the provided MemoryThreshold add 1 second to the counter
			if memStats.Alloc > uint64(b.IndexingSettings.MemoryThresholdMB*MB) {
				secondsAboveTheTresholdCounter++
			} else {
				secondsAboveTheTresholdCounter = 0
			}

			// check if the chunk reader has finished processing
			eof := r.EOF()

			// after 3 second of being above the MemoryThreshold dump to file or after the chunkreader has finished
			if secondsAboveTheTresholdCounter < MemoryThresholdHitsBeforeDumping && !eof {
				continue
			}

			// reset the counter before initiating the dump
			secondsAboveTheTresholdCounter = 0

			// if the chunk reader has finished, also wait for all the workers befor dumping
			if eof {
				err := workers.Wait()
				if err != nil {
					return err
				}
			}

			// perform the dump to file
			if err := b.encode(path); err != nil {
				return err
			}

			// if every document has been read exit
			if eof {
				return nil
			}
		}
	})

	// for every chunk
	for chunk, err := range r.Read() {
		if err != nil {
			return err
		}

		// start a worker
		workers.Go(func() error {

			// for every document
			for _, doc := range chunk {

				// get all the tokens and extract the frequencies
				tokens := preprocess.Tokens(doc.Content, b.PreprocessSettings)
				freqs := preprocess.Frequencies(tokens)

				// add the document and its information to the builder
				b.add(freqs, inverseindex.NewDocument(doc.No, len(doc.Content)))
			}
			return nil
		})
	}

	return dumper.Wait()
}

// Merge merges the information about each term from each partition
//
// inside this function
//   - l prefix stands for local (to a partition)
//   - g prefix stands for global
func Merge(path string) error {
	log.Println("merging...")
	defer log.Println("...end")

	// get all the pa
	partitions, err := ReadPartitions(path)
	if err != nil {
		return err
	}

	if len(partitions) == 0 {
		return nil
	}

	// global term information file
	gTermInfoF, err := os.Create(filepath.Join(path, "terms-info.bin"))
	if err != nil {
		return err
	}
	defer gTermInfoF.Close()

	// global term index file
	gTermF, err := os.Create(filepath.Join(path, "terms.bin"))
	if err != nil {
		return err
	}
	defer gTermF.Close()

	termsEnc := gob.NewEncoder(gTermF)

	gLexicon := iradix.New[uint32]().Txn()
	cw := countwriter.NewWriter(gTermInfoF)

	// for each partition
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

		// create a local lexicon
		lLexicon := iradix.New[uint32]()
		if err := radix.Decode(termF, &lLexicon); err != nil {
			return err
		}

		// append values to gLexicon extracting the information from the local one
		for key, lOffset := range radix.Values(lLexicon) {

			// read a term local information from the local term information file
			lTerm, err := inverseindex.DecodeTerm[inverseindex.LocalTerm](lOffset, lTermInfoR)
			if err != nil {
				return err
			}

			// read the offset of the global term
			gOffset, ok := gLexicon.Get(key)

			// if present
			if ok {

				gTerm, err := inverseindex.DecodeTerm[inverseindex.GlobalTerm](gOffset, gTermInfoF)
				if err != nil {
					return err
				}

				gTerm.DocumentFrequency += lTerm.DocumentFrequency
				gTerm.MaxTermFrequency = max(gTerm.MaxTermFrequency, lTerm.MaxTermFrequency)

				if err := inverseindex.UpdateTerm(gTerm, gOffset, gTermInfoF); err != nil {
					return err
				}
				continue
			}

			// if not already present in the global lexicon

			// create a global Term
			gTerm := &withkey.WithKey[inverseindex.GlobalTerm]{
				Key: string(key),
				Value: inverseindex.GlobalTerm{
					DocumentFrequency: lTerm.DocumentFrequency,
					MaxTermFrequency:  lTerm.MaxTermFrequency,
				},
			}

			// add its pointer to the global lexicon using the count writer
			gLexicon.Insert(key, uint32(cw.Count))

			// add its information to the global term information file (using the count writer)
			if err := inverseindex.EncodeTerm(gTerm, termsEnc, cw); err != nil {
				return err
			}
		}

	}

	return nil
}
