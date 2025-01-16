package seeker

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"

	"github.com/just-hms/pulse/pkg/spimi/inverseindex"
	"github.com/just-hms/pulse/pkg/structures/withkey"
)

type Seeker struct {
	postings, freqs io.ReaderAt

	curOffset uint32
	buf       [4]byte

	DocumentID    uint32
	TermFrequency uint32

	Term withkey.WithKey[inverseindex.LocalTerm]
}

func NewSeeker(postings, freqs io.ReaderAt, t withkey.WithKey[inverseindex.LocalTerm]) *Seeker {
	s := &Seeker{
		Term:      t,
		curOffset: 0,
		postings:  postings,
		freqs:     freqs,
	}
	return s
}

func (s *Seeker) Next() error {
	if EOD(s) {
		return errors.New("cannot read further")
	}

	ptr := s.Term.Value.StartOffset + s.curOffset

	if _, err := s.postings.ReadAt(s.buf[:], int64(ptr)); err != nil {
		return fmt.Errorf("error reading postingFile: %v", err)
	}
	s.DocumentID = binary.LittleEndian.Uint32(s.buf[:])

	if _, err := s.freqs.ReadAt(s.buf[:], int64(ptr)); err != nil {
		return fmt.Errorf("error reading freqFile: %v", err)
	}
	s.TermFrequency = binary.LittleEndian.Uint32(s.buf[:])
	s.curOffset += 4
	return nil
}

func (s *Seeker) String() string {
	return fmt.Sprintf("{%s %d:%d:%d}", s.Term.Key, s.Term.Value.StartOffset, s.curOffset, s.Term.Value.EndOffset)
}

// EOD (end of docs) returns true if the seeker has reached the end of the term's document
func EOD(s *Seeker) bool {
	return s.Term.Value.StartOffset+s.curOffset >= s.Term.Value.EndOffset
}
