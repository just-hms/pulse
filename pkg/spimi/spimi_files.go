package spimi

import (
	"io"
	"path/filepath"

	"github.com/just-hms/pulse/pkg/spimi/inverseindex"
	"golang.org/x/exp/mmap"
)

type SpimiReaders struct {
	inverseindex.LexiconReaders
	DocReader       io.ReaderAt
	TermsInfoReader io.ReaderAt
}

func OpenSpimiFiles(path string) (*SpimiReaders, error) {
	lxReader, err := inverseindex.OpenLexicon(path)
	if err != nil {
		return nil, err
	}

	docReader, err := mmap.Open(filepath.Join(path, "doc.bin"))
	if err != nil {
		return nil, err
	}

	lTermInfoReader, err := mmap.Open(filepath.Join(path, "terms-info.bin"))
	if err != nil {
		return nil, err
	}

	return &SpimiReaders{
		LexiconReaders:  lxReader,
		DocReader:       docReader,
		TermsInfoReader: lTermInfoReader,
	}, nil
}
