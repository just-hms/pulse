package engine

import (
	"io"
	"os"
	"path/filepath"

	iradix "github.com/hashicorp/go-immutable-radix/v2"
	"github.com/just-hms/pulse/pkg/spimi"
	"github.com/just-hms/pulse/pkg/structures/radix"
	"golang.org/x/exp/mmap"
	"golang.org/x/sync/errgroup"
)

type partition struct {
	lLexicon *iradix.Tree[uint32]
	reader   *spimi.SpimiReaders
}

// engine contains all the pre-loaded data structures and file descriptors needed to perform a query
type engine struct {
	path       string
	spimiStats *spimi.Stats

	gLexicon        *iradix.Tree[uint32]
	gTermInfoReader io.ReaderAt

	partitions []*partition
}

// Load return an engine given a path
func Load(path string) (*engine, error) {
	statsFile, err := os.Open(filepath.Join(path, "stats.bin"))
	if err != nil {
		return nil, err
	}

	// spimi settings are loaded, not user provided
	// if stemming is used by the used during indexing it must be used also during querying
	stats, err := spimi.LoadSettings(statsFile)
	if err != nil {
		return nil, err
	}

	gTermsF, err := os.Open(filepath.Join(path, "terms.bin"))
	if err != nil {
		return nil, err
	}
	gLexicon := iradix.New[uint32]()
	if err := radix.Decode(gTermsF, &gLexicon); err != nil {
		return nil, err
	}

	gTermInfoReader, err := mmap.Open(filepath.Join(path, "terms-info.bin"))
	if err != nil {
		return nil, err
	}

	partitionsF, err := spimi.ReadPartitions(path)
	if err != nil {
		return nil, err
	}

	partitions := make([]*partition, len(partitionsF))
	var wg errgroup.Group

	for i, f := range partitionsF {

		// 	load each partition local indexes
		wg.Go(func() error {

			partitionName := filepath.Join(path, f.Name())

			reader, err := spimi.OpenSpimiReaders(partitionName)
			if err != nil {
				return err
			}

			// load each localLexicon for each partition
			lLexicon := iradix.New[uint32]()
			if err := radix.Decode(reader.LexiconReaders.Terms, &lLexicon); err != nil {
				return err
			}

			partitions[i] = &partition{
				lLexicon: lLexicon,
				reader:   reader,
			}
			return nil
		})
	}

	if err := wg.Wait(); err != nil {
		return nil, err
	}

	return &engine{
		path:            path,
		gLexicon:        gLexicon,
		partitions:      partitions,
		spimiStats:      stats,
		gTermInfoReader: gTermInfoReader,
	}, nil
}
