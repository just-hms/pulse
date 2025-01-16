package inverseindex

import (
	"bufio"
	"encoding/binary"
	"encoding/gob"
	"io"
	"iter"
	"slices"

	"maps"

	"github.com/just-hms/pulse/pkg/compression/unary"
	"github.com/just-hms/pulse/pkg/structures/withkey"
)

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
		lx.MaxTermFrequency = max(lx.MaxTermFrequency, frequence)
		lx.DocumentFrequency += 1
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

func (l Lexicon) Encode(loc LexiconFiles, compression bool) error {
	terms := slices.Collect(l.Terms())

	termEnc := gob.NewEncoder(loc.Terms)

	freqEnc := bufio.NewWriter(loc.Freqs)
	defer freqEnc.Flush()

	postEnc := bufio.NewWriter(loc.Posting)
	defer postEnc.Flush()

	var (
		freqStart = uint32(0)
		postStart = uint32(0)
	)

	for _, term := range terms {
		lx := l[term]

		freqLength, err := encodeFrequencies(lx, freqEnc, compression)
		if err != nil {
			return err
		}

		postLength, err := encodePosting(lx, postEnc, compression)
		if err != nil {
			return err
		}

		t := withkey.WithKey[LocalTerm]{
			Key: term,
			Value: LocalTerm{
				GlobalTerm: lx.GlobalTerm,
				FreqStart:  freqStart,
				FreqLength: uint32(freqLength),

				PostStart:  postStart,
				PostLength: uint32(postLength),
			},
		}

		// encode term
		if err := termEnc.Encode(t); err != nil {
			return err
		}

		freqStart += uint32(freqLength)
		postStart += uint32(postLength)
	}

	return nil
}

func encodeFrequencies(lx *LexVal, w io.Writer, compression bool) (int, error) {
	if compression {
		uw := unary.NewWriter(w)
		defer uw.Flush()
		w = uw
	}

	buf := [4]byte{}

	written := 0
	for _, p := range lx.Frequencies {
		binary.LittleEndian.PutUint32(buf[:], p)
		n, err := w.Write(buf[:])
		if err != nil {
			return 0, err
		}
		written += n
	}

	return written, nil
}

func encodePosting(lx *LexVal, w io.Writer, _ bool) (int, error) {
	buf := [4]byte{}

	written := 0
	for _, p := range lx.Posting {
		binary.LittleEndian.PutUint32(buf[:], p)
		n, err := w.Write(buf[:])
		if err != nil {
			return 0, err
		}
		written += n
	}

	return written, nil
}
