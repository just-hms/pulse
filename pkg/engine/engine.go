package engine

import (
	"encoding/gob"
	"errors"
	"fmt"
	"io"
	"io/fs"
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

type DocInfo struct {
	Document inverseindex.Document
	Score    float64
	ID       uint32
}

func MinDoc(a, b *DocInfo) int { return int(a.Score) - int(b.Score) }

func getDocScore(seekers []*seeker.Seeker) *DocInfo {
	res := 0.0
	docID := seekers[0].ID
	// todo: start with tfidf
	for _, s := range seekers {
		res += float64(s.Freq)
	}
	return &DocInfo{Score: res, ID: docID}
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

func Search(query string, path string, k int) ([]*DocInfo, error) {
	qTokens := preprocess.GetTokens(query)

	partitions, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	termsFile, err := os.Open(filepath.Join(path, "terms.bin"))
	if err != nil {
		return nil, err
	}

	globalLexicon, err := getLexicon[inverseindex.GlobalTerm](termsFile)
	if err != nil {
		return nil, err
	}

	qGlobalTerms := make([]withkey.WithKey[inverseindex.GlobalTerm], 0)
	for _, qToken := range qTokens {
		if t, ok := globalLexicon.Get([]byte(qToken)); ok {
			qGlobalTerms = append(qGlobalTerms, withkey.WithKey[inverseindex.GlobalTerm]{
				Key:   qToken,
				Value: t,
			})
		}
	}

	var wg errgroup.Group
	results := make([][]*DocInfo, len(partitions))

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

			qLocalTerms := make([]withkey.WithKey[inverseindex.LocalTerm], 0)
			for _, qTerm := range qGlobalTerms {
				if t, ok := localLexicon.Get([]byte(qTerm.Key)); ok {
					qLocalTerms = append(qLocalTerms, withkey.WithKey[inverseindex.LocalTerm]{
						Key:   qTerm.Key,
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
			seekers := make([]*seeker.Seeker, 0, len(qLocalTerms))
			for _, t := range qLocalTerms {
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
			results[i] = scores.Values()
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

	res := scores.Values()

	err = filldocs(partitions, path, res)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func filldocs(partitions []fs.DirEntry, path string, toClone []*DocInfo) error {
	statsFile, err := os.Open(filepath.Join(path, "stats.bin"))
	if err != nil {
		return err
	}
	defer statsFile.Close()

	cfg, err := config.Load(statsFile)
	if err != nil {
		return err
	}
	docInfos := slices.Clone(toClone)
	slices.SortFunc(docInfos, func(a, b *DocInfo) int { return int(a.ID) - int(b.ID) })

	i := 0
	docsByPartition := make([][]*DocInfo, len(partitions))
	for _, doc := range docInfos {
		// todo: test off by one
		if doc.ID <= cfg.Partitions[i] {
			docsByPartition[i] = append(docsByPartition[i], doc)
		} else {
			i++
		}
	}

	for i, partitionDocs := range docsByPartition {

		if !partitions[i].IsDir() {
			continue
		}

		docsFile, err := os.Open(fmt.Sprintf("%s/doc.bin", filepath.Join(path, partitions[i].Name())))
		if err != nil {
			return err
		}
		for _, doc := range partitionDocs {
			if err := doc.Document.Decode(doc.ID, docsFile); err != nil {
				return err
			}
		}
	}
	return nil
}
