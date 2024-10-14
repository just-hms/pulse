package engine

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"

	"github.com/just-hms/pulse/pkg/engine/config"
	"github.com/just-hms/pulse/pkg/engine/seeker"
	"github.com/just-hms/pulse/pkg/preprocess"
	"github.com/just-hms/pulse/pkg/spimi/inverseindex"
	"github.com/just-hms/pulse/pkg/structures/box"
	"github.com/just-hms/pulse/pkg/structures/radix"
	"github.com/just-hms/pulse/pkg/structures/slicex"
	"github.com/just-hms/pulse/pkg/structures/withkey"
	"golang.org/x/sync/errgroup"
)

func getDocScore(seekers []*seeker.Seeker) *DocInfo {
	res := 0.0
	docID := seekers[0].ID
	// todo: start with tfidf
	for _, s := range seekers {
		res += float64(s.Frequence)
	}
	return &DocInfo{Score: res, ID: docID}
}

func Search(query string, path string, k int) ([]*DocInfo, error) {
	qTokens := preprocess.GetTokens(query)

	globalTermsFile, err := os.Open(filepath.Join(path, "terms.bin"))
	if err != nil {
		return nil, err
	}

	globalLexicon := radix.New[inverseindex.GlobalTerm]()
	err = globalLexicon.Decode(globalTermsFile)
	if err != nil {
		return nil, err
	}

	qGlobalTerms := make([]withkey.WithKey[inverseindex.GlobalTerm], 0)
	for _, qToken := range qTokens {
		if t, ok := globalLexicon.Get([]byte(qToken)); ok {
			qGlobalTerms = append(qGlobalTerms, withkey.WithKey[inverseindex.GlobalTerm]{
				Key:   qToken,
				Value: *t,
			})
		}
	}

	var wg errgroup.Group

	wg.SetLimit(1)

	partitions, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	results := make([][]*DocInfo, len(partitions))

	// 	launch the query for each partition
	for i, partition := range partitions {

		if !partition.IsDir() {
			continue
		}

		wg.Go(func() error {

			folder := filepath.Join(path, partition.Name())

			f, err := inverseindex.OpenLexiconFiles(folder)
			if err != nil {
				return err
			}
			defer f.Close()

			localLexicon := radix.New[inverseindex.LocalTerm]()
			err = localLexicon.Decode(f.TermsFile)
			if err != nil {
				return err
			}

			qLocalTerms := make([]withkey.WithKey[inverseindex.LocalTerm], 0)
			for _, qTerm := range qGlobalTerms {
				if t, ok := localLexicon.Get([]byte(qTerm.Key)); ok {
					qLocalTerms = append(qLocalTerms, withkey.WithKey[inverseindex.LocalTerm]{
						Key:   qTerm.Key,
						Value: *t,
					})
				}
			}

			seekers := make([]*seeker.Seeker, 0, len(qLocalTerms))
			for _, t := range qLocalTerms {
				s := seeker.NewSeeker(f.PostingFile, f.FreqsFile, t)
				s.Next()
				seekers = append(seekers, s)
			}

			scores := box.NewBox(k, func(a, b *DocInfo) int { return a.Less(b) })

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

	scores := box.NewBox(k, func(a, b *DocInfo) int { return a.Less(b) })
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
			if err := doc.Decode(doc.ID, docsFile); err != nil {
				return err
			}
		}
	}
	return nil
}
