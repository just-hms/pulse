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

	counter     uint32
	dumpCounter uint32
}

func NewBuilder() *builder {
	return &builder{
		Lexicon: make(inverseindex.Lexicon),
	}
}

func (s *builder) Lock() {
	s.mu.Lock()
}

func (s *builder) Unlock() {
	s.mu.Unlock()
}

func (s *builder) Add(freqs map[string]uint32, doc inverseindex.Document) {
	doc.ID = s.counter
	s.counter++

	s.Collection = append(s.Collection, doc)

	for term, freq := range freqs {
		if _, ok := s.Lexicon[term]; !ok {
			s.Lexicon[term] = &inverseindex.LexVal{
				Posting:     make([]uint32, 0, 100),
				Frequencies: make([]uint32, 0, 100),
			}
		}

		lx := s.Lexicon[term]
		lx.DocFreq += freq
		lx.Posting = append(lx.Posting, doc.ID)
		lx.Frequencies = append(lx.Frequencies, freq)
	}
}

func (s *builder) Encode(folderPath string) error {
	log.Println("dumping...")
	s.dumpCounter++

	var wg errgroup.Group

	folderPath = path.Join(folderPath, fmt.Sprint(s.dumpCounter))
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
		s.Collection.Encode(docFile)
		return docFile.Close()
	})

	wg.Go(func() error {
		termFile, err := os.Create(path.Join(folderPath, "terms.bin"))
		if err != nil {
			return err
		}
		s.Lexicon.EncodeTerms(termFile)
		return termFile.Close()
	})

	wg.Go(func() error {
		postingFile, err := os.Create(path.Join(folderPath, "posting.bin"))
		if err != nil {
			return err
		}
		s.Lexicon.EncodePostings(postingFile)
		return postingFile.Close()
	})

	wg.Go(func() error {
		freqFile, err := os.Create(path.Join(folderPath, "freqs.bin"))
		if err != nil {
			return err
		}
		s.Lexicon.EncodeTerms(freqFile)
		return freqFile.Close()
	})

	if err := wg.Wait(); err != nil {
		return err
	}
	// make this better
	log.Println("..end")
	s.Lexicon.Clear()
	s.Collection.Clear()
	return nil
}
