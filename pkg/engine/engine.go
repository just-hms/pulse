package engine

import (
	"encoding/gob"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"

	"github.com/just-hms/pulse/pkg/spimi/inverseindex"
	"github.com/just-hms/pulse/pkg/spimi/preprocess"
	"golang.org/x/sync/errgroup"
)

func Search(q string, path string, k int) ([]uint32, error) {
	tokens := preprocess.GetTokens(q)

	partitions, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}
	var wg errgroup.Group

	type seeker struct {
		id    uint32
		score uint32
		seeek int64
		stop  int64
	}

	type docInfo struct {
		score uint32
		id    uint32
	}

	results := make([][]docInfo, len(partitions))

	// 	launch the query for each partition
	for i, partition := range partitions {

		i := i
		partition := partition

		wg.Go(func() error {

			terms := make([]inverseindex.Term, 0, len(tokens))

			termsFile, err := os.Open(fmt.Sprintf("%s/terms.bin", filepath.Join(path, partition.Name())))
			if err != nil {
				return err
			}

			termDecoder := gob.NewDecoder(termsFile)
			// read the terms start and shit
			for {
				t := inverseindex.Term{}
				err := termDecoder.Decode(&t)
				if err == io.EOF {
					break
				}
				if err != nil {
					return err
				}

				if slices.Contains(tokens, t.Value) {
					terms = append(terms, t)
				}
			}

			docFile, err := os.Open(fmt.Sprintf("%s/doc.bin", filepath.Join(path, partition.Name())))
			if err != nil {
				return err
			}
			docDecoder := gob.NewDecoder(termsFile)

			freqFile, err := os.Open(fmt.Sprintf("%s/freqs.bin", filepath.Join(path, partition.Name())))
			if err != nil {
				return err
			}
			freqDecoder := gob.NewDecoder(freqFile)

			seekers := make([]*seeker, len(terms))

			for i, t := range terms {
				curSeek := &seeker{}
				curSeek.seeek = int64(t.StartOffset)

				var id, freq uint32
				docFile.Seek(curSeek.seeek, 0)
				docDecoder.Decode(&id)
				freqFile.Seek(curSeek.seeek, 0)
				freqDecoder.Decode(&freq)

				curSeek.id = id
				// TODO: do the actual score
				curSeek.score = freq
				curSeek.seeek += 32 * 8
				seekers[i] = curSeek
			}

			scores := make([]docInfo, 0, k)

			for {
				if len(seekers) == 0 {
					break
				}
				curSeek := slices.MinFunc(seekers, func(a, b *seeker) int {
					return int(a.id) - int(b.id)
				})

				// ADD TO THE RESULTS
				scores = append(scores, docInfo{
					id:    curSeek.id,
					score: curSeek.score,
				})
				slices.SortFunc(scores, func(a, b docInfo) int {
					return int(a.score) - int(b.score)
				})
				// keep only k elements
				scores = scores[:min(len(scores), k)]

				// SEEK TO THE NEXT

				var id, freq uint32
				docFile.Seek(curSeek.seeek, 0)
				docDecoder.Decode(&id)
				freqFile.Seek(curSeek.seeek, 0)
				freqDecoder.Decode(&freq)

				curSeek.id = id
				// TODO: do the actual score
				curSeek.score = freq
				curSeek.seeek += 32 * 8

				// TODO: remove finished seekers
				seekers = slices.DeleteFunc(seekers, func(s *seeker) bool {
					return s.seeek > s.stop
				})
			}
			results[i] = scores
			fmt.Println(scores)
			return nil
		})
	}

	if err := wg.Wait(); err != nil {
		return []uint32{}, err
	}

	var scores []docInfo
	for _, res := range results {
		scores = append(scores, res...)
	}

	slices.SortFunc(scores, func(a, b docInfo) int {
		return int(a.score) - int(b.score)
	})
	res := make([]uint32, k)
	for i := range scores {
		res[i] = scores[i].id
	}

	// 	unisci le liste
	// 	return the first k results
	return res, nil
}
