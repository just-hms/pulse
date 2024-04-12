package inverseindex

import (
	"bufio"
	"encoding/binary"
	"encoding/gob"
	"io"

	"github.com/just-hms/pulse/pkg/structures/withkey"
	"golang.org/x/exp/maps"
)

type Lexicon map[string]*LexVal

type LexVal struct {
	DocFreq     uint32
	Posting     []uint32
	Frequencies []uint32
}

func (l Lexicon) Clear() {
	maps.Clear(l)
}

func (l Lexicon) Terms() []string {
	return maps.Keys(l)
}

func (l Lexicon) EncodeTerms(w io.Writer, terms []string) error {
	enc := gob.NewEncoder(w)
	var cur uint32 = 0

	for _, term := range terms {
		lx := l[term]

		// TODO: specify the bit position in future
		span := uint32(len(lx.Posting)) * 4
		t := withkey.WithKey[LocalTerm]{
			Key: term,
			Value: LocalTerm{
				GlobalTerm: GlobalTerm{
					DocFreq: lx.DocFreq,
				},
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
		buf := make([]byte, 4)
		for _, p := range lx.Posting {
			binary.LittleEndian.PutUint32(buf, p)
			_, err := enc.Write(buf)
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
		buf := make([]byte, 4)
		for _, p := range lx.Frequencies {
			binary.LittleEndian.PutUint32(buf, p)
			_, err := enc.Write(buf)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
