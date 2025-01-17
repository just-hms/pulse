package engine

import (
	"io"
	"math"
	"os"
	"path/filepath"
	"slices"

	iradix "github.com/hashicorp/go-immutable-radix/v2"
	"github.com/just-hms/pulse/pkg/engine/seeker"
	"github.com/just-hms/pulse/pkg/preprocess"
	"github.com/just-hms/pulse/pkg/spimi"
	"github.com/just-hms/pulse/pkg/spimi/inverseindex"
	"github.com/just-hms/pulse/pkg/spimi/stats"
	"github.com/just-hms/pulse/pkg/structures/heap"
	"github.com/just-hms/pulse/pkg/structures/radix"
	"github.com/just-hms/pulse/pkg/structures/slicex"
	"github.com/just-hms/pulse/pkg/structures/withkey"
	"golang.org/x/exp/mmap"
	"golang.org/x/sync/errgroup"
)

type engine struct {
	path string

	gLexicon        *iradix.Tree[uint32]
	gTermInfoReader io.ReaderAt

	readers   []*spimi.SpimiReaders
	lLexicons []*iradix.Tree[uint32]

	stats *stats.Stats
}

func score(doc *DocInfo, seekers []*seeker.Seeker, stats *stats.Stats, s *Settings) {
	switch s.Metric {
	case TFIDF:
		score := 0.0
		for _, s := range seekers {
			TF := float64(s.TermFrequency)
			IDF := math.Log(float64(stats.N) / float64(s.Term.Value.DocumentFrequency))
			score += (1 + math.Log(TF)) * IDF
		}
		doc.Score = score
	case BM25:
		score := 0.0
		for _, s := range seekers {
			TF := float64(s.TermFrequency)
			IDF := math.Log(float64(stats.N) / float64(s.Term.Value.DocumentFrequency))
			score += TF / (BM25_k1*((1-BM25_b)+BM25_b*(float64(doc.Size)/stats.ADL)) + TF) * IDF
		}
		doc.Score = score
	default:
		panic("metric not implemented")
	}
}

func Load(path string) (*engine, error) {
	statsFile, err := os.Open(filepath.Join(path, "stats.bin"))
	if err != nil {
		return nil, err
	}

	stats, err := stats.Load(statsFile)
	if err != nil {
		return nil, err
	}

	gTermsF, err := os.Open(filepath.Join(path, "terms.bin"))
	if err != nil {
		return nil, err
	}
	gLexicon := iradix.New[uint32]()
	if err := radix.Decode(gTermsF, &gLexicon); err != nil {
		return nil, err
	}

	gTermInfoReader, err := mmap.Open(filepath.Join(path, "terms-info.bin"))
	if err != nil {
		return nil, err
	}

	partitions, err := spimi.ReadPartitions(path)
	if err != nil {
		return nil, err
	}

	readers := make([]*spimi.SpimiReaders, len(partitions))
	lLexicons := make([]*iradix.Tree[uint32], len(partitions))

	var wg errgroup.Group

	for i, partition := range partitions {
		// 	launch the query for each partition
		wg.Go(func() error {
			partitionName := filepath.Join(path, partition.Name())
			reader, err := spimi.OpenSpimiFiles(partitionName)
			if err != nil {
				return err
			}

			lLexicon := iradix.New[uint32]()
			if err := radix.Decode(reader.LexiconReaders.Terms, &lLexicon); err != nil {
				return err
			}

			lLexicons[i] = lLexicon
			readers[i] = reader
			return nil
		})
	}

	if err := wg.Wait(); err != nil {
		return nil, err
	}

	return &engine{
		path:            path,
		gLexicon:        gLexicon,
		readers:         readers,
		stats:           stats,
		lLexicons:       lLexicons,
		gTermInfoReader: gTermInfoReader,
	}, nil
}

func (e *engine) Search(query string, s *Settings) ([]*DocInfo, error) {
	qTokens := preprocess.GetTokens(query, e.stats.IndexingSettings)

	qgTerms := make([]withkey.WithKey[inverseindex.GlobalTerm], 0)
	for _, qToken := range qTokens {
		if gOffset, ok := e.gLexicon.Get([]byte(qToken)); ok {
			gTerm, err := inverseindex.DecodeTerm[inverseindex.GlobalTerm](gOffset, e.gTermInfoReader)
			if err != nil {
				return nil, err
			}

			qgTerms = append(qgTerms, withkey.WithKey[inverseindex.GlobalTerm]{
				Key:   qToken,
				Value: *gTerm,
			})
		}
	}

	results := make([]heap.Heap[*DocInfo], len(e.readers))
	var wg errgroup.Group
	for i := range e.readers {
		// 	launch the query for each partition
		wg.Go(func() error {
			results[i].Cap = s.K
			return e.searchPartition(i, qgTerms, e.stats, s, &results[i])
		})
	}

	if err := wg.Wait(); err != nil {
		return nil, err
	}

	result := &heap.Heap[*DocInfo]{Cap: s.K}
	for _, res := range results {
		result.Push(res.Values()...)
	}
	vals := result.Values()
	slices.Reverse(vals)
	return vals, nil
}

func (e *engine) searchPartition(i int, qgTerms []withkey.WithKey[inverseindex.GlobalTerm], stats *stats.Stats, s *Settings, result *heap.Heap[*DocInfo]) error {
	readers := e.readers[i]

	qlTerms := make([]withkey.WithKey[inverseindex.LocalTerm], 0)
	for _, qgTerm := range qgTerms {
		lOffest, ok := e.lLexicons[i].Get([]byte(qgTerm.Key))
		if ok {
			lTerm, err := inverseindex.DecodeTerm[inverseindex.LocalTerm](lOffest, readers.TermsInfo)
			if err != nil {
				return err
			}

			qlTerms = append(qlTerms, withkey.WithKey[inverseindex.LocalTerm]{
				Key: qgTerm.Key,
				Value: inverseindex.LocalTerm{
					GlobalTerm: qgTerm.Value,
					FreqStart:  lTerm.FreqStart,
					FreqLength: lTerm.FreqLength,
					PostStart:  lTerm.PostStart,
					PostLength: lTerm.FreqLength,
				},
			})
			continue
		}

		// if a term is not present in any document and the query is conjuctive return 0 results
		if s.Conjunctive {
			return nil
		}
	}

	seekers := make([]*seeker.Seeker, 0, len(qlTerms))
	for _, qlTerm := range qlTerms {
		s := seeker.NewSeeker(readers.Posting, readers.Freqs, qlTerm, stats.Compression)
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

		// if the current document does not contain all the terms go to the next
		if len(curSeeks) != len(qlTerms) && s.Conjunctive {
			for _, s := range curSeeks {
				s.Next()
			}
			// remove all finished seekers
			seekers = slices.DeleteFunc(seekers, seeker.EOD)
			continue
		}

		doc := &DocInfo{
			ID: curSeeks[0].DocumentID,
		}
		if err := doc.Decode(doc.ID, readers.DocReader); err != nil {
			return err
		}

		score(doc, curSeeks, stats, s)
		result.Push(doc)

		// seek to the next
		for _, s := range curSeeks {
			s.Next()
		}
		seekers = slices.DeleteFunc(seekers, seeker.EOD)
	}
	return nil
}
