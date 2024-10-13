package engine

import (
	"encoding/gob"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"

	iradix "github.com/hashicorp/go-immutable-radix/v2"
	"github.com/just-hms/pulse/pkg/engine/config"
	"github.com/just-hms/pulse/pkg/engine/seeker"
	"github.com/just-hms/pulse/pkg/preprocess"
	"github.com/just-hms/pulse/pkg/spimi/inverseindex"
	"github.com/just-hms/pulse/pkg/structures/box"
	"github.com/just-hms/pulse/pkg/structures/slicex"
	"github.com/just-hms/pulse/pkg/structures/withkey"
	"golang.org/x/sync/errgroup"
)

type docInfo struct {
	score float64
	id    uint32
}

func MinDoc(a, b docInfo) int { return int(a.score) - int(b.score) }

func getDocScore(seekers []*seeker.Seeker) docInfo {
	res := 0.0
	docID := seekers[0].ID
	// TODO: start with tfidf
	for _, s := range seekers {
		res += float64(s.Freq)
	}
	return docInfo{score: res, id: docID}
}

func getLexicon[T any](reader io.Reader) (*iradix.Tree[T], error) {

	lexicon := iradix.New[T]().Txn()

	termDecoder := gob.NewDecoder(reader)
	// read the terms start and shit
	for {
		t := withkey.WithKey[T]{}
		err := termDecoder.Decode(&t)
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, err
		}

		lexicon.Insert([]byte(t.Key), t.Value)
	}

	return lexicon.Commit(), nil
}

func Search(q string, path string, k int) (inverseindex.Collection, error) {
	tokens := preprocess.GetTokens(q)

	partitions, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}
	var wg errgroup.Group

	termsFile, err := os.Open(filepath.Join(path, "terms.bin"))
	if err != nil {
		return nil, err
	}

	globalLexicon, err := getLexicon[inverseindex.GlobalTerm](termsFile)
	if err != nil {
		return nil, err
	}

	globalTerms := make([]withkey.WithKey[inverseindex.GlobalTerm], 0)
	for _, token := range tokens {
		if t, ok := globalLexicon.Get([]byte(token)); ok {
			globalTerms = append(globalTerms, withkey.WithKey[inverseindex.GlobalTerm]{
				Key:   token,
				Value: t,
			})
		}
	}

	stats, err := config.Load(path)
	if err != nil {
		return nil, err
	}

	fmt.Println(stats)

	results := make([][]docInfo, len(partitions))

	// 	launch the query for each partition
	for i, partition := range partitions {

		if !partition.IsDir() {
			continue
		}

		wg.Go(func() error {

			folder := filepath.Join(path, partition.Name())

			termsFile, err := os.Open(filepath.Join(folder, "terms.bin"))
			if err != nil {
				return err
			}

			localLexicon, err := getLexicon[inverseindex.LocalTerm](termsFile)
			if err != nil {
				return err
			}

			localTerms := make([]withkey.WithKey[inverseindex.LocalTerm], 0)
			for _, gTerm := range globalTerms {
				if t, ok := localLexicon.Get([]byte(gTerm.Key)); ok {
					localTerms = append(localTerms, withkey.WithKey[inverseindex.LocalTerm]{
						Key:   gTerm.Key,
						Value: t,
					})
				}
			}
			postingsFile, err := os.Open(filepath.Join(folder, "posting.bin"))
			if err != nil {
				return err
			}

			freqFile, err := os.Open(filepath.Join(folder, "freqs.bin"))
			if err != nil {
				return err
			}
			seekers := make([]*seeker.Seeker, 0, len(localTerms))
			for _, t := range localTerms {
				s := seeker.NewSeeker(postingsFile, freqFile, t)
				s.Next()
				seekers = append(seekers, s)
			}

			scores := box.NewBox(MinDoc, k)

			for {
				if len(seekers) == 0 {
					break
				}
				curSeeks := slicex.MinsFunc(seekers, func(a, b *seeker.Seeker) int {
					return int(a.ID) - int(b.ID)
				})

				score := getDocScore(curSeeks)
				scores.Add(score)

				// seek to the next
				for _, s := range curSeeks {
					s.Next()
				}
				// remove all finished seekers
				seekers = slices.DeleteFunc(seekers, seeker.EOD)
			}
			results[i] = scores.List()
			return nil
		})
	}

	if err := wg.Wait(); err != nil {
		return nil, err
	}

	scores := box.NewBox(MinDoc, k)
	for _, res := range results {
		scores.Add(res...)
	}

	resIDs := make([]uint32, 0)
	for _, doc := range scores.List() {
		resIDs = append(resIDs, doc.id)
	}

	docs := make(inverseindex.Collection, 0, len(tokens))

	// TODO: access docs by ID (which in incremental)
	for _, partition := range partitions {

		if !partition.IsDir() {
			continue
		}

		docsFile, err := os.Open(fmt.Sprintf("%s/doc.bin", filepath.Join(path, partition.Name())))
		if err != nil {
			return nil, err
		}

		docsDecoder := gob.NewDecoder(docsFile)
		// read the terms start and shit
		for {
			doc := inverseindex.Document{}
			err := docsDecoder.Decode(&doc)
			if errors.Is(err, io.EOF) {
				break
			}
			if err != nil {
				return nil, err
			}

			if slices.Contains(resIDs, doc.ID) {
				docs = append(docs, doc)
			}
		}
	}

	return docs, nil
}
