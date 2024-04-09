package engine

import (
	"encoding/gob"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"

	"github.com/just-hms/pulse/pkg/engine/seeker"
	"github.com/just-hms/pulse/pkg/preprocess"
	"github.com/just-hms/pulse/pkg/spimi/inverseindex"
	"github.com/just-hms/pulse/pkg/structures/slicex"
	"golang.org/x/sync/errgroup"
)

type docInfo struct {
	score float64
	id    uint32
}

func getDocScore(seekers []*seeker.Seeker) docInfo {
	res := 0.0
	docID := seekers[0].ID
	for _, s := range seekers {
		res += float64(s.Freq)
	}
	return docInfo{score: res, id: docID}
}

func Search(q string, path string, k int) ([]uint32, error) {
	tokens := preprocess.GetTokens(q)

	partitions, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}
	var wg errgroup.Group

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
				if errors.Is(err, io.EOF) {
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

			freqFile, err := os.Open(fmt.Sprintf("%s/freqs.bin", filepath.Join(path, partition.Name())))
			if err != nil {
				return err
			}

			seekers := make([]*seeker.Seeker, 0, len(terms))
			for _, t := range terms {
				s := seeker.NewSeeker(docFile, freqFile, t)
				s.Next()
				seekers = append(seekers, s)
			}

			scores := make([]docInfo, 0, k)

			for {
				if len(seekers) == 0 {
					break
				}
				curSeeks := slicex.MinsFunc(seekers, func(a, b *seeker.Seeker) int {
					return int(a.ID) - int(b.ID)
				})

				docScore := getDocScore(curSeeks)

				// TODO: refactor
				scores = append(scores, docScore)
				slices.SortFunc(scores, func(a, b docInfo) int {
					return int(a.score) - int(b.score)
				})
				// keep only k elements
				scores = slicex.Cap(scores, k)

				// seek to the next
				for _, s := range curSeeks {
					s.Next()
				}
				// remove all finished seekers
				seekers = slices.DeleteFunc(seekers, seeker.EOD)
			}
			results[i] = scores
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
	scores = slicex.Cap(scores, k)
	res := make([]uint32, len(scores))
	for i := range scores {
		res[i] = scores[i].id
	}

	// 	unisci le liste
	// 	return the first k results
	return res, nil
}
