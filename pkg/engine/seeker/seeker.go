package seeker

import (
	"encoding/binary"
	"fmt"
	"io"
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

	// Cache for faster processing
	postingBuffer []byte
	freqBuffer    []byte
}

func NewSeeker(postingFile, freqFile *os.File, t withkey.WithKey[inverseindex.LocalTerm]) *Seeker {
	// Define a reasonable buffer size (adjust as needed)
	bufferSize := int(t.Value.EndOffset) - int(t.Value.StartOffset)

	s := &Seeker{
		Term:          t,
		start:         t.Value.StartOffset,
		cur:           0,
		end:           t.Value.EndOffset,
		postingFile:   postingFile,
		freqFile:      freqFile,
		postingBuffer: make([]byte, bufferSize),
		freqBuffer:    make([]byte, bufferSize),
	}
	s.fill()
	return s
}

// Read next chunk of data into buffer
func (s *Seeker) fill() error {
	var err error
	if _, err = s.postingFile.Seek(int64(s.cur), io.SeekStart); err != nil {
		return fmt.Errorf("error seeking postingFile: %v", err)
	}

	_, err = s.postingFile.Read(s.postingBuffer)
	if err != nil && err != io.EOF {
		return fmt.Errorf("error reading postingFile: %v", err)
	}

	if _, err = s.freqFile.Seek(int64(s.cur), io.SeekStart); err != nil {
		return fmt.Errorf("error seeking freqFile: %v", err)
	}

	_, err = s.freqFile.Read(s.freqBuffer)
	if err != nil && err != io.EOF {
		return fmt.Errorf("error reading freqFile: %v", err)
	}

	return nil
}

func (s *Seeker) Next() error {
	s.DocumentID = binary.LittleEndian.Uint32(s.postingBuffer[s.cur : s.cur+4])
	s.Frequence = binary.LittleEndian.Uint32(s.freqBuffer[s.cur : s.cur+4])
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
