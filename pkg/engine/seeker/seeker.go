package seeker

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"

	"github.com/just-hms/pulse/pkg/compression/deltavarint"
	"github.com/just-hms/pulse/pkg/compression/unary"
	"github.com/just-hms/pulse/pkg/spimi/inverseindex"
	"github.com/just-hms/pulse/pkg/structures/withkey"
)

// Sekeer is the data structure used to travers both the posting list and the frequencies of a given term
type Seeker struct {
	postings, freqs io.Reader

	buf [4]byte
	eod bool // the seeker arrived to the end of the documents

	DocumentID    uint32
	TermFrequency uint32

	Term withkey.WithKey[inverseindex.LocalTerm]
}

func NewSeeker(postings, freqs io.ReaderAt, t withkey.WithKey[inverseindex.LocalTerm], compression bool) *Seeker {
	// a term contains the start and the end of the two lists
	// create a NewSectionReader for both to use the Read() call instead of the ReadAt()
	// this simplifies performing sequential reads.
	s := &Seeker{
		Term:     t,
		postings: io.NewSectionReader(postings, int64(t.Value.PostStart), int64(t.Value.PostLength)),
		freqs:    io.NewSectionReader(freqs, int64(t.Value.FreqStart), int64(t.Value.PostLength)),
	}
	// if compression is enable use unary compression for the frequencies and delta variable bytes for the posting list
	if compression {
		s.freqs = unary.NewReader(s.freqs, 1)
		s.postings = deltavarint.NewReader(s.postings)
	}
	return s
}

// Next reads the next posting list/frencuency setting the s.DocumentID and s.TermFrequency fields inside the Seeker struct
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

// String return the string representation of the Seeker
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
