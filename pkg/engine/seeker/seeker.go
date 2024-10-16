package seeker

import (
	"encoding/binary"
	"fmt"
	"os"

	"github.com/just-hms/pulse/pkg/spimi/inverseindex"
	"github.com/just-hms/pulse/pkg/structures/withkey"
)

type Seeker struct {
	postingFile, freqFile *os.File

	DocumentID      uint32
	Frequence       uint32
	start, cur, end uint32

	Term withkey.WithKey[inverseindex.LocalTerm]
}

func NewSeeker(postingFile, freqFile *os.File, t withkey.WithKey[inverseindex.LocalTerm]) *Seeker {
	return &Seeker{
		Term:  t,
		start: t.Value.StartOffset,
		cur:   t.Value.StartOffset,
		end:   t.Value.EndOffset,

		postingFile: postingFile,
		freqFile:    freqFile,
	}
}

func (s *Seeker) Next() {
	if _, err := s.postingFile.Seek(int64(s.cur), 0); err != nil {
		panic(fmt.Sprintf("seeker: %v, while reading docs, with error: %v", s, err))
	}
	if err := binary.Read(s.postingFile, binary.LittleEndian, &s.DocumentID); err != nil {
		panic(fmt.Sprintf("seeker: %v, while reading docs, with error: %v", s, err))
	}

	if _, err := s.freqFile.Seek(int64(s.cur), 0); err != nil {
		panic(fmt.Sprintf("seeker: %v, while reading docs, with error: %v", s, err))
	}

	if err := binary.Read(s.freqFile, binary.LittleEndian, &s.Frequence); err != nil {
		panic(fmt.Sprintf("seeker: %v, while reading freqs, with error: %v", s, err))
	}

	// todo: specify bit position when using compression
	s.cur += 4
}

func (s *Seeker) String() string {
	return fmt.Sprintf("{%s %d:%d:%d}", s.Term.Key, s.start, s.cur, s.end)
}

// EOD (end of docs) returns true if the seeker have reached the end of the term's document
func EOD(s *Seeker) bool {
	return s.cur > s.end
}
