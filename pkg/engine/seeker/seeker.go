package seeker

import (
	"encoding/binary"
	"fmt"
	"os"

	"github.com/just-hms/pulse/pkg/spimi/inverseindex"
)

type Seeker struct {
	docFile, freqFile *os.File

	ID              uint32
	Freq            uint32
	start, pos, end int64

	// for debugging only
	term string
}

func NewSeeker(docFile, freqFile *os.File, t inverseindex.Term) *Seeker {
	return &Seeker{
		term:     t.Value,
		start:    int64(t.StartOffset),
		pos:      int64(t.StartOffset),
		end:      int64(t.EndOffset),
		docFile:  docFile,
		freqFile: freqFile,
	}
}

func (s *Seeker) Next() {
	s.docFile.Seek(s.pos, 0)
	err := binary.Read(s.docFile, binary.LittleEndian, &s.ID)
	if err != nil {
		panic(fmt.Sprintf("seeker failed %v with error: %v", s, err))
	}
	s.freqFile.Seek(s.pos, 0)
	err = binary.Read(s.freqFile, binary.LittleEndian, &s.Freq)
	if err != nil {
		panic(fmt.Sprintf("seeker failed %v with error: %v", s, err))
	}
	// TODO: specify bit position in future
	s.pos += 4
}

func (s *Seeker) String() string {
	return fmt.Sprintf("{%s:%d\\in[%d,%d]}", s.term, s.pos, s.start, s.end)

}

// EOD end of docs, return true if the seekers have seeked to the end of the term's documents
func EOD(s *Seeker) bool {
	return s.pos > s.end
}