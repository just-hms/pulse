package seeker

import (
	"encoding/binary"
	"fmt"
	"io"

	"github.com/just-hms/pulse/pkg/bufseekio"
	"github.com/just-hms/pulse/pkg/spimi/inverseindex"
	"github.com/just-hms/pulse/pkg/structures/withkey"
)

type Seeker struct {
	postingFile, freqFile io.ReadSeeker

	DocumentID      uint32
	Frequence       uint32
	start, cur, end uint32

	Term withkey.WithKey[inverseindex.LocalTerm]
}

func NewSeeker(postingFile, freqFile io.ReadSeeker, t withkey.WithKey[inverseindex.LocalTerm]) *Seeker {
	bufferSize := int(t.Value.EndOffset) - int(t.Value.StartOffset)
	s := &Seeker{
		Term:        t,
		start:       t.Value.StartOffset,
		cur:         0,
		end:         t.Value.EndOffset,
		postingFile: bufseekio.NewReaderSize(postingFile, bufferSize),
		freqFile:    bufseekio.NewReaderSize(freqFile, bufferSize),
	}
	return s
}

func (s *Seeker) Next() error {
	var err error

	if _, err = s.postingFile.Seek(int64(s.cur), io.SeekStart); err != nil {
		return fmt.Errorf("error seeking postingFile: %v", err)
	}
	if err = binary.Read(s.postingFile, binary.LittleEndian, &s.DocumentID); err != nil {
		return fmt.Errorf("error reading postingFile: %v", err)
	}
	if _, err = s.freqFile.Seek(int64(s.cur), io.SeekStart); err != nil {
		return fmt.Errorf("error seeking freqFile: %v", err)
	}
	if err = binary.Read(s.freqFile, binary.LittleEndian, &s.Frequence); err != nil {
		return fmt.Errorf("error reading freqFile: %v", err)
	}
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
