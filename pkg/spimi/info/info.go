package info

import (
	"encoding/gob"
	"fmt"
	"log"
	"os"
	"path"
	"slices"
	"sync"

	"golang.org/x/exp/maps"
)

type spimiBuilder struct {
	Lexicon Lexicon
	Docs    []Doc
	mu      sync.Mutex

	counter     uint32
	dumpCounter uint32
}

func NewSpimiBuilder() *spimiBuilder {
	return &spimiBuilder{
		Lexicon: make(Lexicon),
	}
}

func (s *spimiBuilder) Lock() {
	s.mu.Lock()
}

func (s *spimiBuilder) Unlock() {
	s.mu.Unlock()
}

func (s *spimiBuilder) Add(freqs map[string]uint32, doc Doc) {
	doc.id = s.counter
	s.counter++

	s.Docs = append(s.Docs, doc)

	for term, freq := range freqs {
		if _, ok := s.Lexicon[term]; !ok {
			s.Lexicon[term] = &lexVal{
				Posting:     make([]uint32, 0, 100),
				Frequencies: make([]uint32, 0, 100),
			}
		}

		lx := s.Lexicon[term]
		lx.DocFreq += freq
		lx.Posting = append(lx.Posting, doc.id)
		lx.Frequencies = append(lx.Frequencies, freq)
	}
}

func (s *spimiBuilder) Dump(folderPath string) error {
	log.Println("dumping...")
	s.dumpCounter++

	var wg sync.WaitGroup

	folderPath = path.Join(folderPath, fmt.Sprint(s.dumpCounter))
	err := os.MkdirAll(folderPath, 0o755)
	if err != nil {
		return err
	}

	wg.Add(1)
	go func() {
		defer wg.Done()

		// encode the doc
		docFile, err := os.Create(path.Join(folderPath, "doc.bin"))
		if err != nil {
			return
		}
		docE := gob.NewEncoder(docFile)
		for _, doc := range s.Docs {
			docE.Encode(doc.id)
			docE.Encode(doc.No)
			docE.Encode(doc.Len)
		}
		docFile.Close()
	}()

	terms := maps.Keys(s.Lexicon)
	slices.Sort(terms)

	wg.Add(1)
	go func() {
		defer wg.Done()
		lexFile, err := os.Create(path.Join(folderPath, "lex.bin"))
		if err != nil {
			return
		}
		lexE := gob.NewEncoder(lexFile)
		var cur uint32 = 0
		for _, term := range terms {
			lx := s.Lexicon[term]

			span := uint32(len(lx.Posting)) * 32 * 8
			t := Term{
				Value:       term,
				DocFreq:     lx.DocFreq,
				StartOffset: cur,
				EndOffset:   cur + span,
			}

			lexE.Encode(t)
			cur += span
		}
		lexFile.Close()
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		postingFile, err := os.Create(path.Join(folderPath, "posting.bin"))
		if err != nil {
			return
		}
		postingE := gob.NewEncoder(postingFile)
		for _, term := range terms {
			lx := s.Lexicon[term]
			postingE.Encode(lx.Posting)
		}
		postingFile.Close()
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		freqFile, err := os.Create(path.Join(folderPath, "freqs.bin"))
		if err != nil {
			return
		}
		freqE := gob.NewEncoder(freqFile)
		for _, term := range terms {
			lx := s.Lexicon[term]
			freqE.Encode(lx.Frequencies)
		}
		freqFile.Close()
	}()
	wg.Wait()

	maps.Clear(s.Lexicon)
	s.Docs = nil
	s.Docs = make([]Doc, 0)

	log.Println("..end")
	return nil
}

type Term struct {
	Value       string
	DocFreq     uint32
	StartOffset uint32
	EndOffset   uint32
}

type Doc struct {
	id  uint32
	No  string
	Len int
}

type Lexicon map[string]*lexVal

type lexVal struct {
	DocFreq     uint32
	Posting     []uint32
	Frequencies []uint32
}
