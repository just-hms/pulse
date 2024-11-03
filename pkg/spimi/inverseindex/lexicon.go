package inverseindex

import (
	"bufio"
	"encoding/binary"
	"encoding/gob"
	"io"
	"iter"
	"slices"

	"maps"

	"github.com/just-hms/pulse/pkg/structures/withkey"
	"golang.org/x/sync/errgroup"
)

// todo: maybe use radix ??
type Lexicon map[string]*LexVal

type LexVal struct {
	GlobalTerm
	Posting     []uint32
	Frequencies []uint32
}

func (l Lexicon) Add(freqs map[string]uint32, docID uint32) {
	for term, frequence := range freqs {
		if _, ok := l[term]; !ok {
			l[term] = &LexVal{
				Posting:     make([]uint32, 0, 100),
				Frequencies: make([]uint32, 0, 100),
			}
		}
		lx := l[term]
		lx.Frequence += frequence
		lx.MaxDocFrequence = max(lx.MaxDocFrequence, frequence)
		lx.Appearences += 1
		lx.Posting = append(lx.Posting, docID)
		lx.Frequencies = append(lx.Frequencies, frequence)
	}
}

func (l Lexicon) Clear() {
	clear(l)
}

func (l Lexicon) Terms() iter.Seq[string] {
	return maps.Keys(l)
}

func (l Lexicon) Encode(loc LexiconFiles) error {
	var wg errgroup.Group

	terms := slices.Collect(l.Terms())

	wg.Go(func() error {
		return l.EncodeTerms(loc.Terms, terms)
	})

	wg.Go(func() error {
		return l.EncodePostings(loc.Posting, terms)
	})

	wg.Go(func() error {
		return l.EncodeFreqs(loc.Freqs, terms)
	})

	return wg.Wait()
}

func (l Lexicon) EncodeTerms(w io.Writer, terms []string) error {
	enc := gob.NewEncoder(w)
	var cur uint32 = 0

	for _, term := range terms {
		lx := l[term]

		// todo: specify the bit position in future
		span := uint32(len(lx.Posting)) * 4
		t := withkey.WithKey[LocalTerm]{
			Key: term,
			Value: LocalTerm{
				GlobalTerm:  lx.GlobalTerm,
				StartOffset: cur,
				EndOffset:   cur + span,
			},
		}

		if err := enc.Encode(t); err != nil {
			return err
		}
		cur += span
	}
	return nil
}

func (l Lexicon) EncodePostings(w io.Writer, terms []string) error {
	enc := bufio.NewWriter(w)
	defer enc.Flush()

	for _, term := range terms {
		lx := l[term]
		for _, p := range lx.Posting {
			err := binary.Write(enc, binary.LittleEndian, p)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (l Lexicon) EncodeFreqs(w io.Writer, terms []string) error {
	enc := bufio.NewWriter(w)
	defer enc.Flush()

	for _, term := range terms {
		lx := l[term]
		for _, p := range lx.Frequencies {
			err := binary.Write(enc, binary.LittleEndian, p)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
