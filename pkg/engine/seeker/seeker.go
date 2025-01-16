package seeker

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"

	"github.com/just-hms/pulse/pkg/compression/unary"
	"github.com/just-hms/pulse/pkg/spimi/inverseindex"
	"github.com/just-hms/pulse/pkg/structures/withkey"
)

type Seeker struct {
	postings, freqs io.Reader

	buf [4]byte
	eod bool // the seeker arrived to the end of the documents

	DocumentID    uint32
	TermFrequency uint32

	Term withkey.WithKey[inverseindex.LocalTerm]
}

func NewSeeker(postings, freqs io.ReaderAt, t withkey.WithKey[inverseindex.LocalTerm], compression bool) *Seeker {
	s := &Seeker{
		Term:     t,
		postings: io.NewSectionReader(postings, int64(t.Value.PostStart), int64(t.Value.PostLength)),
		freqs:    io.NewSectionReader(freqs, int64(t.Value.FreqStart), int64(t.Value.PostLength)),
	}
	if compression {
		s.freqs = unary.NewReader(s.freqs)
	}
	return s
}

func (s *Seeker) Next() error {
	if _, err := s.postings.Read(s.buf[:]); err != nil {
		if errors.Is(err, io.EOF) {
			s.eod = true
			return nil
		}
		return fmt.Errorf("error reading postingFile: %w", err)
	}
	s.DocumentID = binary.LittleEndian.Uint32(s.buf[:])

	if _, err := s.freqs.Read(s.buf[:]); err != nil {
		if errors.Is(err, io.EOF) {
			s.eod = true
			return nil
		}
		return fmt.Errorf("error reading freqFile: %w", err)
	}
	s.TermFrequency = binary.LittleEndian.Uint32(s.buf[:])
	return nil
}

func (s *Seeker) String() string {
	return fmt.Sprintf(
		"{%s postings:[%d,%d], terms[%d,%d]}",
		s.Term.Key,
		s.Term.Value.PostStart, s.Term.Value.PostLength,
		s.Term.Value.FreqStart, s.Term.Value.FreqLength,
	)
}

// EOD (end of docs) returns true if the seeker has reached the end of the term's document
func EOD(s *Seeker) bool {
	return s.eod
}
