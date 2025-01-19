package spimi

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"slices"
	"sync"

	"github.com/just-hms/pulse/pkg/spimi/inverseindex"
	"golang.org/x/sync/errgroup"
)

// builder is the object responsible for building the index
type builder struct {
	lexicon    inverseindex.Lexicon
	collection inverseindex.Collection
	mu         sync.Mutex

	documentPartitionCounter uint32
	IndexingSettings
}

// NewBuilder returns a spimi.builder
func NewBuilder(s IndexingSettings) *builder {
	return &builder{
		lexicon:                  inverseindex.Lexicon{},
		collection:               inverseindex.Collection{},
		mu:                       sync.Mutex{},
		documentPartitionCounter: 0,
		IndexingSettings:         s,
	}
}

// add adds a document to the builder
func (b *builder) add(freqs map[string]uint32, doc inverseindex.Document) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.collection.Add(doc)

	// the ID is incremental and starts from 0 for each document partition
	ID := uint32(b.collection.Len() - 1)
	b.lexicon.Add(freqs, ID)
}

// encode is called to dump the builder information into a given folder
func (b *builder) encode(path string) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	defer func() { b.documentPartitionCounter++ }()

	log.Println("dumping...")
	defer log.Println("..end")

	// create a folder for the partition using the current counter as name (0,1,2,3....)
	partitionPath := filepath.Join(path, fmt.Sprint(b.documentPartitionCounter))
	err := os.MkdirAll(partitionPath, 0o755)
	if err != nil {
		return err
	}

	var wg errgroup.Group

	// write down the documents
	wg.Go(func() error {
		f, err := os.Create(filepath.Join(partitionPath, "doc.bin"))
		if err != nil {
			return err
		}
		defer f.Close()
		return b.collection.Encode(f)

	})

	// write down the lexicon
	wg.Go(func() error {
		f, err := inverseindex.CreateLexicon(partitionPath)
		if err != nil {
			return err
		}
		defer f.Close()
		return b.lexicon.Encode(f, b.IndexingSettings.Compression)
	})

	if err := wg.Wait(); err != nil {
		return err
	}

	// write down the current stats
	{

		f, err := os.OpenFile(filepath.Join(path, "stats.bin"), os.O_RDWR|os.O_CREATE, 0600)
		if err != nil {
			return err
		}

		// read the current stats
		s, _ := LoadSettings(f)

		// add the indexSettings to the stats
		s.IndexingSettings = b.IndexingSettings

		// update the stats
		s.UpdateStats(uint32(b.collection.Len()), b.collection.AvgDocumentSize)

		// dump the updated stats into the same file
		if err := s.Dump(f); err != nil {
			return err
		}

		defer f.Close()
	}

	b.clear()
	return nil
}

// clear clears all the builde informations
func (b *builder) clear() {
	b.lexicon.Clear()
	b.collection.Clear()
}

// ReadPartitions returns the file entry of all the partitions
func ReadPartitions(path string) ([]fs.DirEntry, error) {
	partitions, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	return slices.DeleteFunc(partitions, func(p fs.DirEntry) bool { return !p.IsDir() }), nil
}
