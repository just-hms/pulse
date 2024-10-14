package inverseindex

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

type LexiconFiles struct {
	TermsFile, PostingFile, FreqsFile *os.File
}

func OpenLexiconFiles(path string) (LexiconFiles, error) {
	return openLexiconFiles(path, os.O_RDWR|os.O_CREATE)
}

func CreateLexiconFiles(path string) (LexiconFiles, error) {
	return openLexiconFiles(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC)
}

// OpenLexiconFiles opens or creates the necessary lexicon files (terms.bin, posting.bin, freqs.bin)
// at the specified path and returns a LexiconFiles struct with open file handles.
func openLexiconFiles(path string, flag int) (LexiconFiles, error) {
	termsPath := filepath.Join(path, "terms.bin")
	postingPath := filepath.Join(path, "posting.bin")
	freqsPath := filepath.Join(path, "freqs.bin")

	termsFile, err := os.OpenFile(termsPath, flag, 0644)
	if err != nil {
		return LexiconFiles{}, fmt.Errorf("failed to open terms file: %w", err)
	}

	postingFile, err := os.OpenFile(postingPath, flag, 0644)
	if err != nil {
		termsFile.Close()
		return LexiconFiles{}, fmt.Errorf("failed to open posting file: %w", err)
	}

	freqsFile, err := os.OpenFile(freqsPath, flag, 0644)
	if err != nil {
		termsFile.Close()
		postingFile.Close()
		return LexiconFiles{}, fmt.Errorf("failed to open freqs file: %w", err)
	}

	return LexiconFiles{
		TermsFile:   termsFile,
		PostingFile: postingFile,
		FreqsFile:   freqsFile,
	}, nil
}

// Close closes all the open file handles in LexiconFiles.
func (l *LexiconFiles) Close() error {
	err := l.TermsFile.Close()
	err = errors.Join(err, l.PostingFile.Close())
	err = errors.Join(err, l.FreqsFile.Close())
	return err
}
