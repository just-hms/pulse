package spimi

import (
	"fmt"
	"log"
	"os"
	"path"
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

func (b *builder) Encode(folderPath string) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	defer func() { b.dumpCounter++ }()

	log.Println("dumping...")

	var wg errgroup.Group

	folderPath = path.Join(folderPath, fmt.Sprint(b.dumpCounter))
	err := os.MkdirAll(folderPath, 0o755)
	if err != nil {
		return err
	}

	wg.Go(func() error {
		// encode the doc
		docFile, err := os.Create(path.Join(folderPath, "doc.bin"))
		if err != nil {
			return err
		}
		b.Collection.Encode(docFile)
		return docFile.Close()
	})

	// TODO: in future the encode should be done in a single function (the freqs and posting info must be compressed)
	terms := b.Lexicon.Terms()

	wg.Go(func() error {
		termFile, err := os.Create(path.Join(folderPath, "terms.bin"))
		if err != nil {
			return err
		}
		b.Lexicon.EncodeTerms(termFile, terms)
		return termFile.Close()
	})

	wg.Go(func() error {
		postingFile, err := os.Create(path.Join(folderPath, "posting.bin"))
		if err != nil {
			return err
		}
		b.Lexicon.EncodePostings(postingFile, terms)
		return postingFile.Close()
	})

	wg.Go(func() error {
		freqFile, err := os.Create(path.Join(folderPath, "freqs.bin"))
		if err != nil {
			return err
		}
		b.Lexicon.EncodeFreqs(freqFile, terms)
		return freqFile.Close()
	})

	if err := wg.Wait(); err != nil {
		return err
	}
	// make this better
	log.Println("..end")
	b.Lexicon.Clear()
	b.Collection.Clear()
	return nil
}
