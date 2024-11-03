package inverseindex

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"golang.org/x/exp/mmap"
)

type LexiconReaders struct {
	Terms          *os.File
	Posting, Freqs io.ReaderAt
}

func OpenLexicon(path string) (LexiconReaders, error) {
	termsReader, err := os.Open(filepath.Join(path, "terms.bin"))
	if err != nil {
		return LexiconReaders{}, err
	}

	postingReader, err := mmap.Open(filepath.Join(path, "posting.bin"))
	if err != nil {
		return LexiconReaders{}, fmt.Errorf("failed to open posting file: %w", err)
	}

	freqReader, err := mmap.Open(filepath.Join(path, "freqs.bin"))
	if err != nil {
		return LexiconReaders{}, fmt.Errorf("failed to open posting file: %w", err)
	}

	return LexiconReaders{
		Terms:   termsReader,
		Posting: postingReader,
		Freqs:   freqReader,
	}, nil

}

// Close closes all the open file handles in LexiconFiles.
func (l *LexiconReaders) Close() error {
	return l.Terms.Close()
}

type LexiconFiles struct {
	Terms, Posting, Freqs *os.File
}

func CreateLexicon(path string) (LexiconFiles, error) {
	flag := os.O_RDWR | os.O_CREATE | os.O_TRUNC

	termsFile, err := os.OpenFile(filepath.Join(path, "terms.bin"), flag, 0644)
	if err != nil {
		return LexiconFiles{}, fmt.Errorf("failed to open terms file: %w", err)
	}

	postingFile, err := os.OpenFile(filepath.Join(path, "posting.bin"), flag, 0644)
	if err != nil {
		termsFile.Close()
		return LexiconFiles{}, fmt.Errorf("failed to open posting file: %w", err)
	}

	freqsFile, err := os.OpenFile(filepath.Join(path, "freqs.bin"), flag, 0644)
	if err != nil {
		termsFile.Close()
		postingFile.Close()
		return LexiconFiles{}, fmt.Errorf("failed to open freqs file: %w", err)
	}

	return LexiconFiles{
		Terms:   termsFile,
		Posting: postingFile,
		Freqs:   freqsFile,
	}, nil
}

// Close closes all the open file handles in LexiconFiles.
func (l *LexiconFiles) Close() error {
	return l.Terms.Close()
}
