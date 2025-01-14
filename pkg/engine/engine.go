package engine

import (
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
	path          string
	globalLexicon *iradix.Tree[*inverseindex.GlobalTerm]
	localLexicons []*iradix.Tree[*inverseindex.LocalTerm]
	partititions  []string
	stats         *stats.Stats
}

func GetUpperScore(seekers []*seeker.Seeker, s *Settings) float64 {
	switch s.Metric {
	case TFIDF:
		res := 0.0
		for _, s := range seekers {
			res += float64(s.Term.Value.MaxDocFrequence)
		}
		return res
	case BM25:
		// todo: implement upperscore for bm25
		return math.MaxFloat64
	default:
		panic("metric not implemented")
	}
}

func calculateDocInfo(doc *DocInfo, seekers []*seeker.Seeker, stats *stats.Stats, s *Settings) {
	switch s.Metric {
	case TFIDF:
		score := 0.0
		for _, s := range seekers {
			TF := (float64(s.Frequence) / float64(doc.Size))
			IDF := math.Log(float64(stats.CollectionSize) / float64(s.Frequence))
			score += TF * IDF
		}
		doc.Score = score
		break
	case BM25:
		score := 0.0
		for _, s := range seekers {
			IDF := math.Log((float64(stats.CollectionSize)-float64(s.Term.Value.Appearences)+0.5)/(float64(s.Term.Value.Appearences)+0.5) + 1)
			score += IDF * (float64(s.Frequence)*(BM25_k1+1)/float64(s.Frequence) + BM25_k1*(1-BM25_b+BM25_b*(float64(doc.Size)/stats.AverageDocumentSize)))
		}
		doc.Score = score
		break
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

	globalTermsFile, err := os.Open(filepath.Join(path, "terms.bin"))
	if err != nil {
		return nil, err
	}

	globalLexicon := iradix.New[*inverseindex.GlobalTerm]()
	err = radix.Decode(globalTermsFile, &globalLexicon)
	if err != nil {
		return nil, err
	}

	partitions, err := spimi.ReadPartitions(path)
	if err != nil {
		return nil, err
	}

	localLexicons := make([]*iradix.Tree[*inverseindex.LocalTerm], len(partitions))
	partitionsName := make([]string, len(partitions))

	var wg errgroup.Group

	for i, partition := range partitions {
		// 	launch the query for each partition
		wg.Go(func() error {
			partitionsName[i] = filepath.Join(path, partition.Name())
			f, err := inverseindex.OpenLexicon(partitionsName[i])
			if err != nil {
				return err
			}
			defer f.Close()

			localLexicons[i] = iradix.New[*inverseindex.LocalTerm]()
			err = radix.Decode(f.Terms, &localLexicons[i])
			if err != nil {
				return err
			}

			return nil
		})
	}

	if err := wg.Wait(); err != nil {
		return nil, err
	}

	return &engine{
		path:          path,
		globalLexicon: globalLexicon,
		localLexicons: localLexicons,
		partititions:  partitionsName,
		stats:         stats,
	}, nil
}

func (e *engine) Search(query string, s *Settings) ([]*DocInfo, error) {
	qTokens := preprocess.GetTokens(query)

	qGlobalTerms := make([]withkey.WithKey[inverseindex.GlobalTerm], 0)
	for _, qToken := range qTokens {
		if t, ok := e.globalLexicon.Get([]byte(qToken)); ok {
			qGlobalTerms = append(qGlobalTerms, withkey.WithKey[inverseindex.GlobalTerm]{
				Key:   qToken,
				Value: *t,
			})
		}
	}

	results := make([]heap.Heap[*DocInfo], len(e.partititions))
	var wg errgroup.Group

	for i := range e.partititions {
		// 	launch the query for each partition
		wg.Go(func() error {
			return e.searchPartition(i, qGlobalTerms, e.stats, s, &results[i])
		})
	}

	if err := wg.Wait(); err != nil {
		return nil, err
	}

	result := heap.Heap[*DocInfo]{}
	for _, res := range results {
		result.Add(res.Values()...)
	}
	result.Resize(s.K)

	return result.Values(), nil
}

func (e *engine) searchPartition(i int, qGlobalTerms []withkey.WithKey[inverseindex.GlobalTerm], stats *stats.Stats, s *Settings, result *heap.Heap[*DocInfo]) error {
	lexReaders, err := inverseindex.OpenLexicon(e.partititions[i])
	if err != nil {
		return err
	}
	defer lexReaders.Close()

	docReader, err := mmap.Open(filepath.Join(e.partititions[i], "doc.bin"))
	if err != nil {
		return err
	}

	qLocalTerms := make([]withkey.WithKey[inverseindex.LocalTerm], 0)
	for _, qTerm := range qGlobalTerms {
		t, ok := e.localLexicons[i].Get([]byte(qTerm.Key))
		if ok {
			qLocalTerms = append(qLocalTerms, withkey.WithKey[inverseindex.LocalTerm]{
				Key:   qTerm.Key,
				Value: *t,
			})
			continue
		}

		// if a term is not present in any document and the query is conjuctive return 0 results
		if s.Conjunctive {
			return nil
		}
	}

	seekers := make([]*seeker.Seeker, 0, len(qLocalTerms))
	for _, t := range qLocalTerms {
		s := seeker.NewSeeker(lexReaders.Posting, lexReaders.Freqs, t)
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
		if len(curSeeks) != len(qLocalTerms) && s.Conjunctive {
			for _, s := range curSeeks {
				s.Next()
			}
			// remove all finished seekers
			seekers = slices.DeleteFunc(seekers, seeker.EOD)
			continue
		}

		peek, ok := result.Peek()
		if ok && result.Size() >= s.K && GetUpperScore(curSeeks, s) < peek.Score {
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
		if err := doc.Decode(doc.ID, docReader); err != nil {
			return err
		}

		calculateDocInfo(doc, curSeeks, stats, s)
		result.Add(doc)

		if result.Size() >= s.K*2 {
			result.Resize(s.K)
		}

		// seek to the next
		for _, s := range curSeeks {
			s.Next()
		}
		seekers = slices.DeleteFunc(seekers, seeker.EOD)
	}
	result.Resize(s.K)
	return nil
}
