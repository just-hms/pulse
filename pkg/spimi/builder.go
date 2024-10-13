package spimi

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/just-hms/pulse/pkg/spimi/inverseindex"
	"golang.org/x/sync/errgroup"
)

type builder struct {
	Lexicon    inverseindex.Lexicon
	Collection inverseindex.Collection
	mu         sync.Mutex

	docCounter  uint32
	dumpCounter uint32
}

func newBuilder() *builder {
	return &builder{
		Lexicon: make(inverseindex.Lexicon),
	}
}

func (b *builder) Add(freqs map[string]uint32, doc inverseindex.Document) {
	b.mu.Lock()
	defer b.mu.Unlock()

	doc.ID = b.docCounter
	b.docCounter++

	b.Collection = append(b.Collection, doc)

	for term, freq := range freqs {
		if _, ok := b.Lexicon[term]; !ok {
			b.Lexicon[term] = &inverseindex.LexVal{
				Posting:     make([]uint32, 0, 100),
				Frequencies: make([]uint32, 0, 100),
			}
		}

		lx := b.Lexicon[term]
		lx.DocFreq += freq
		lx.Posting = append(lx.Posting, doc.ID)
		lx.Frequencies = append(lx.Frequencies, freq)
	}
}

func (b *builder) Encode(path string) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	defer func() { b.dumpCounter++ }()

	log.Println("dumping...")
	defer log.Println("..end")

	var wg errgroup.Group

	partitionPath := filepath.Join(path, fmt.Sprint(b.dumpCounter))
	err := os.MkdirAll(partitionPath, 0o755)
	if err != nil {
		return err
	}

	wg.Go(func() error {
		// encode the doc
		docFile, err := os.Create(filepath.Join(partitionPath, "doc.bin"))
		if err != nil {
			return err
		}
		b.Collection.Encode(docFile)
		return docFile.Close()
	})

	// TODO: in future the encode should be done in a single function (the freqs and posting info must be compressed)
	terms := b.Lexicon.Terms()

	wg.Go(func() error {
		termFile, err := os.Create(filepath.Join(partitionPath, "terms.bin"))
		if err != nil {
			return err
		}
		b.Lexicon.EncodeTerms(termFile, terms)
		return termFile.Close()
	})

	wg.Go(func() error {
		postingFile, err := os.Create(filepath.Join(partitionPath, "posting.bin"))
		if err != nil {
			return err
		}
		b.Lexicon.EncodePostings(postingFile, terms)
		return postingFile.Close()
	})

	wg.Go(func() error {
		freqFile, err := os.Create(filepath.Join(partitionPath, "freqs.bin"))
		if err != nil {
			return err
		}
		b.Lexicon.EncodeFreqs(freqFile, terms)
		return freqFile.Close()
	})

	if err := wg.Wait(); err != nil {
		return err
	}

	statsFile, err := os.OpenFile(filepath.Join(path, "stats.bin"), os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return err
	}

	_, err = statsFile.WriteString(fmt.Sprintf("%d\n", b.docCounter))
	if err != nil {
		return err
	}

	defer statsFile.Close()

	// make this better
	b.Lexicon.Clear()
	b.Collection.Clear()
	return nil
}
