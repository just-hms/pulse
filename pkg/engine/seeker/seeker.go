package seeker

import (
	"encoding/binary"
	"fmt"
	"os"

	"github.com/just-hms/pulse/pkg/spimi/inverseindex"
	"github.com/just-hms/pulse/pkg/structures/withkey"
)

type Seeker struct {
	docFile, freqFile *os.File

	ID              uint32
	Freq            uint32
	start, pos, end uint32

	// debugging purpose, will probably be removed
	term string
}

func NewSeeker(docFile, freqFile *os.File, t withkey.WithKey[inverseindex.LocalTerm]) *Seeker {
	return &Seeker{
		term:     t.Key,
		start:    t.Value.StartOffset,
		pos:      t.Value.StartOffset,
		end:      t.Value.EndOffset,
		docFile:  docFile,
		freqFile: freqFile,
	}
}

func (s *Seeker) Next() {
	s.docFile.Seek(int64(s.pos), 0)
	if err := binary.Read(s.docFile, binary.LittleEndian, &s.ID); err != nil {
		panic(fmt.Sprintf("seeker: %v, while reading docs, with error: %v", s, err))
	}

	s.freqFile.Seek(int64(s.pos), 0)
	if err := binary.Read(s.freqFile, binary.LittleEndian, &s.Freq); err != nil {
		panic(fmt.Sprintf("seeker: %v, while reading freqs, with error: %v", s, err))
	}

	// : specify bit position in future
	s.pos += 4
}

func (s *Seeker) String() string {
	return fmt.Sprintf("{%s %d:%d:%d}", s.term, s.start, s.pos, s.end)
}

// EOD (end of docs) returns true if the seeker have reached the end of the term's document
func EOD(s *Seeker) bool {
	return s.pos > s.end
}
