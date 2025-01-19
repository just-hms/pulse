package engine

import (
	"math"
	"slices"

	"github.com/just-hms/pulse/pkg/engine/seeker"
	"github.com/just-hms/pulse/pkg/preprocess"
	"github.com/just-hms/pulse/pkg/spimi"
	"github.com/just-hms/pulse/pkg/spimi/inverseindex"
	"github.com/just-hms/pulse/pkg/structures/heap"
	"github.com/just-hms/pulse/pkg/structures/slicex"
	"github.com/just-hms/pulse/pkg/structures/withkey"
	"golang.org/x/sync/errgroup"
)

// Search given a query and a set of settings returns the top k documents
//
// inside this function
//   - q prefix stands for query
//   - l prefix stands for local (to a partition)
//   - q prefix stands for global
func (e *engine) Search(query string, seachSettings *SearchSettings) ([]*DocumentInfo, error) {

	// extract query tokens from the query using the same preprocessing settings used during indexing
	qTokens := preprocess.Tokens(query, e.spimiStats.PreprocessSettings)

	// load each term global information
	qgTerms := make([]withkey.WithKey[inverseindex.GlobalTerm], 0)
	for _, qToken := range qTokens {

		if gOffset, ok := e.gLexicon.Get([]byte(qToken)); ok {

			// decode the term
			gTerm, err := inverseindex.DecodeTerm[inverseindex.GlobalTerm](gOffset, e.gTermInfoReader)
			if err != nil {
				return nil, err
			}

			// append the term to the least of queried terms
			qgTerms = append(qgTerms, withkey.WithKey[inverseindex.GlobalTerm]{
				Key:   qToken,
				Value: *gTerm,
			})
		}
	}

	// create a result for each partition
	results := make([]heap.Heap[*DocumentInfo], len(e.partitions))
	var wg errgroup.Group

	// launch the query in each partition
	for i := range e.partitions {
		wg.Go(func() error {
			results[i].Cap = seachSettings.K
			return e.searchPartition(i, qgTerms, seachSettings, &results[i])
		})
	}

	if err := wg.Wait(); err != nil {
		return nil, err
	}

	// add each partition's result to the final result
	result := heap.Heap[*DocumentInfo]{Cap: seachSettings.K}
	for _, res := range results {
		result.Push(res.Values()...)
	}

	// reverse the result ranking (document are store in increasing Score order to leave the smallest as root)
	vals := result.Values()
	slices.Reverse(vals)

	return vals, nil
}

// searchPartition executes a query inside a partition and fills up the result
//
// inside this function
//   - q prefix stands for query
//   - l prefix stands for local (to a partition)
//   - q prefix stands for global
func (e *engine) searchPartition(i int, qgTerms []withkey.WithKey[inverseindex.GlobalTerm], searchSettings *SearchSettings, result *heap.Heap[*DocumentInfo]) error {
	var (
		reader   = e.partitions[i].reader
		lLexicon = e.partitions[i].lLexicon
	)

	qlTerms := make([]withkey.WithKey[inverseindex.LocalTerm], 0)
	for _, qgTerm := range qgTerms {

		// load the local term information
		lOffest, ok := lLexicon.Get([]byte(qgTerm.Key))
		if ok {
			lTerm, err := inverseindex.DecodeTerm[inverseindex.LocalTerm](lOffest, reader.TermsInfo)
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
					PostLength: lTerm.PostLength,
				},
			})
			continue
		}

		// if a term is not present in any document and the query is conjuctive return 0 results
		if searchSettings.Conjunctive {
			return nil
		}
	}

	// DAAT create a seeker for each searched term
	seekers := make([]*seeker.Seeker, 0, len(qlTerms))
	for _, qlTerm := range qlTerms {
		s := seeker.NewSeeker(reader.Posting, reader.Freqs, qlTerm, e.spimiStats.Compression)
		seekers = append(seekers, s)
	}

	// load one document into each seeker
	updateSeekers(seekers)

	// continue until all the seekers finished while removing all the finished seekers at each iteration

	for ; len(seekers) != 0; seekers = slices.DeleteFunc(seekers, seeker.EOD) {
		// the current seeker involve the smallest document ID
		curSeeks := slicex.MinsFunc(seekers, func(a, b *seeker.Seeker) int {
			return int(a.DocumentID) - int(b.DocumentID)
		})

		// if the current document does not contain all the terms and the query mode is Conjunctive go to the next document
		if len(curSeeks) != len(qlTerms) && searchSettings.Conjunctive {
			updateSeekers(curSeeks)
			continue
		}

		// extract the current document id from the first seekers (all current seekers have the same docID)
		curDocID := curSeeks[0].DocumentID

		doc := &inverseindex.Document{}
		if err := doc.Decode(curDocID, reader.DocReader); err != nil {
			return err
		}

		// add the document into the result heap
		result.Push(&DocumentInfo{
			ID:       curDocID,
			Document: doc,
			Score:    score(doc, curSeeks, e.spimiStats, searchSettings.Metric),
		})

		updateSeekers(curSeeks)
	}

	return nil
}

func updateSeekers(seekers []*seeker.Seeker) {
	for _, s := range seekers {
		s.Next()
	}
}

func score(doc *inverseindex.Document, seekers []*seeker.Seeker, stats *spimi.Stats, metric Metric) float64 {
	score := 0.0

	switch metric {
	case TFIDF:
		for _, s := range seekers {
			TF := float64(s.TermFrequency)
			IDF := math.Log(float64(stats.N) / float64(s.Term.Value.DocumentFrequency))
			score += (1 + math.Log(TF)) * IDF
		}
		return score
	case BM25:
		for _, s := range seekers {
			TF := float64(s.TermFrequency)
			IDF := math.Log(float64(stats.N) / float64(s.Term.Value.DocumentFrequency))
			score += TF / (BM25_k1*((1-BM25_b)+BM25_b*(float64(doc.Size)/stats.ADL)) + TF) * IDF
		}
		return score
	default:
		panic("metric not implemented")
	}
}
