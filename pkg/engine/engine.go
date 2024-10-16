package engine

import (
	"io/fs"
	"math"
	"os"
	"path/filepath"
	"slices"

	iradix "github.com/hashicorp/go-immutable-radix/v2"
	"github.com/just-hms/pulse/pkg/engine/seeker"
	"github.com/just-hms/pulse/pkg/preprocess"
	"github.com/just-hms/pulse/pkg/spimi/inverseindex"
	"github.com/just-hms/pulse/pkg/spimi/stats"
	"github.com/just-hms/pulse/pkg/structures/box"
	"github.com/just-hms/pulse/pkg/structures/radix"
	"github.com/just-hms/pulse/pkg/structures/slicex"
	"github.com/just-hms/pulse/pkg/structures/withkey"
	"golang.org/x/sync/errgroup"
)

// todo: put this as a input of the search
var metric Metric

func CalculateDocInfo(seekers []*seeker.Seeker, doc inverseindex.Document, stats *stats.Stats) *DocInfo {
	documentID := seekers[0].DocumentID

	switch metric {
	case TFIDF:
		score := 0.0
		for _, s := range seekers {
			TF := (float64(s.Frequence) / float64(doc.Size))
			IDF := math.Log(float64(stats.CollectionSize) / float64(s.Frequence))
			score += TF * IDF
		}
		return &DocInfo{Score: score, Document: doc, ID: documentID}
	case BM25:
		score := 0.0
		for _, s := range seekers {
			IDF := math.Log((float64(stats.CollectionSize)-float64(s.Term.Value.Appearences)+0.5)/(float64(s.Term.Value.Appearences)+0.5) + 1)
			score += IDF * (float64(s.Frequence)*(BM25_k1+1)/float64(s.Frequence) + BM25_k1*(1-BM25_b+BM25_b*(float64(doc.Size)/stats.AverageDocumentSize)))
		}
		return &DocInfo{Score: score, Document: doc, ID: documentID}
	default:
		panic("metric not implemented")
	}
}

func Search(query string, path string, s *Settings) ([]*DocInfo, error) {
	qTokens := preprocess.GetTokens(query)

	statsFile, err := os.Open(filepath.Join(path, "stats.bin"))
	if err != nil {
		return nil, err
	}

	stats, err := stats.Load(statsFile)
	if err != nil {
		return nil, err
	}

	globalTermsFile, err := os.Open(filepath.Join(path, "terms.bin"))
	if err != nil {
		return nil, err
	}

	globalLexicon := iradix.New[*inverseindex.GlobalTerm]()
	err = radix.Decode(globalTermsFile, &globalLexicon)
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

	partitions, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	partitions = slices.DeleteFunc(partitions, func(p fs.DirEntry) bool { return !p.IsDir() })

	results := make([]box.Box[*DocInfo], len(partitions))
	var wg errgroup.Group

	for i, partition := range partitions {
		// 	launch the query for each partition
		wg.Go(func() error {
			results[i] = box.NewBox(s.K, func(a, b *DocInfo) int { return a.More(b) })
			folder := filepath.Join(path, partition.Name())
			return searchPartition(folder, qGlobalTerms, stats, results[i])
		})
	}

	if err := wg.Wait(); err != nil {
		return nil, err
	}

	result := box.NewBox(s.K, func(a, b *DocInfo) int { return a.More(b) })
	for _, res := range results {
		result.Add(res.Values()...)
	}

	return result.Values(), nil
}

func searchPartition(path string, qGlobalTerms []withkey.WithKey[inverseindex.GlobalTerm], stats *stats.Stats, result box.Box[*DocInfo]) error {
	f, err := inverseindex.OpenLexiconFiles(path)
	if err != nil {
		return err
	}
	defer f.Close()

	docsFile, err := os.Open(filepath.Join(path, "doc.bin"))
	if err != nil {
		return err
	}

	localLexicon := iradix.New[*inverseindex.LocalTerm]()
	err = radix.Decode(f.TermsFile, &localLexicon)
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

	for {
		if len(seekers) == 0 {
			break
		}
		curSeeks := slicex.MinsFunc(seekers, func(a, b *seeker.Seeker) int {
			return int(a.DocumentID) - int(b.DocumentID)
		})

		doc := inverseindex.Document{}
		if err := doc.Decode(curSeeks[0].DocumentID, docsFile); err != nil {
			return err
		}

		docInfo := CalculateDocInfo(curSeeks, doc, stats)
		result.Add(docInfo)

		// seek to the next
		for _, s := range curSeeks {
			s.Next()
		}
		// remove all finished seekers
		seekers = slices.DeleteFunc(seekers, seeker.EOD)
	}

	return nil
}
