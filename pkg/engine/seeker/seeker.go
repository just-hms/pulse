package seeker

import (
	"encoding/binary"
	"fmt"
	"io"

	"github.com/just-hms/pulse/pkg/spimi/inverseindex"
	"github.com/just-hms/pulse/pkg/structures/withkey"
)

type Seeker struct {
	postings, freqs io.ReaderAt

	// todo: test []byte
	buf [4]byte

	DocumentID      uint32
	Frequence       uint32
	start, cur, end uint32

	Term withkey.WithKey[inverseindex.LocalTerm]
}

func NewSeeker(postings, freqs io.ReaderAt, t withkey.WithKey[inverseindex.LocalTerm]) *Seeker {
	s := &Seeker{
		Term:     t,
		start:    t.Value.StartOffset,
		cur:      0,
		end:      t.Value.EndOffset,
		postings: postings,
		freqs:    freqs,
	}
	return s
}

func (s *Seeker) Next() error {
	if _, err := s.postings.ReadAt(s.buf[:], int64(s.cur)*4); err != nil {
		return fmt.Errorf("error reading postingFile: %v", err)
	}
	s.DocumentID = binary.LittleEndian.Uint32(s.buf[:])

	if _, err := s.freqs.ReadAt(s.buf[:], int64(s.cur)*4); err != nil {
		return fmt.Errorf("error reading freqFile: %v", err)
	}
	s.Frequence = binary.LittleEndian.Uint32(s.buf[:])
	s.cur += 4
	return nil
}

func (s *Seeker) String() string {
	return fmt.Sprintf("{%s %d:%d:%d}", s.Term.Key, s.start, s.cur, s.end)
}

// EOD (end of docs) returns true if the seeker has reached the end of the term's document
func EOD(s *Seeker) bool {
	return s.start+s.cur >= s.end
}
