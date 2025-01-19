package inverseindex

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"golang.org/x/exp/mmap"
)

// LexiconReaders data structure containing all the lexicon file descriptors for reading
type LexiconReaders struct {
	Terms                     *os.File
	TermsInfo, Posting, Freqs io.ReaderAt
}

// OpenLexicon opens all the LexiconReaders given a path to a folder
func OpenLexicon(path string) (LexiconReaders, error) {
	termsReader, err := os.Open(filepath.Join(path, "terms.bin"))
	if err != nil {
		return LexiconReaders{}, err
	}

	termsInfoReader, err := mmap.Open(filepath.Join(path, "terms-info.bin"))
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
		Terms:     termsReader,
		Posting:   postingReader,
		Freqs:     freqReader,
		TermsInfo: termsInfoReader,
	}, nil

}

// Close closes all the open file handles in LexiconFiles.
func (l *LexiconReaders) Close() error {
	return l.Terms.Close()
}

// LexiconFiles data structure containing all the lexicon file descriptors for writing
type LexiconFiles struct {
	Terms, Posting, Freqs, TermsInfo *os.File
}

// CreateLexicon creates all the LexiconFiles given a path to an existing folder
func CreateLexicon(path string) (LexiconFiles, error) {
	flag := os.O_RDWR | os.O_CREATE | os.O_TRUNC

	termsFile, err := os.OpenFile(filepath.Join(path, "terms.bin"), flag, 0644)
	if err != nil {
		return LexiconFiles{}, fmt.Errorf("failed to open terms file: %w", err)
	}

	termsInfoFile, err := os.OpenFile(filepath.Join(path, "terms-info.bin"), flag, 0644)
	if err != nil {
		return LexiconFiles{}, fmt.Errorf("failed to open terms-info file: %w", err)
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
		Terms:     termsFile,
		Posting:   postingFile,
		Freqs:     freqsFile,
		TermsInfo: termsInfoFile,
	}, nil
}

// Close closes all the open file handles in LexiconFiles.
func (l *LexiconFiles) Close() error {
	return l.Terms.Close()
}
