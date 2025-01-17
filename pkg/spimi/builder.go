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
	"github.com/just-hms/pulse/pkg/spimi/stats"
	"golang.org/x/sync/errgroup"
)

type builder struct {
	lexicon    inverseindex.Lexicon
	collection inverseindex.Collection
	mu         sync.Mutex

	dumpCounter uint32
	stats.IndexingSettings
}

func NewBuilder(s stats.IndexingSettings) *builder {
	return &builder{
		lexicon:          inverseindex.Lexicon{},
		collection:       inverseindex.Collection{},
		mu:               sync.Mutex{},
		dumpCounter:      0,
		IndexingSettings: s,
	}
}

func (b *builder) add(freqs map[string]uint32, doc inverseindex.Document) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.collection.Add(doc)
	b.lexicon.Add(freqs, uint32(b.collection.Len()-1))
}

func (b *builder) encode(path string) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	defer func() { b.dumpCounter++ }()

	log.Println("dumping...")
	defer log.Println("..end")

	var wg errgroup.Group

	partitionPath := filepath.Join(path, fmt.Sprint(b.dumpCounter))
	err := os.MkdirAll(partitionPath, 0o755)
	if err != nil {
		return err
	}

	wg.Go(func() error {
		f, err := os.Create(filepath.Join(partitionPath, "doc.bin"))
		if err != nil {
			return err
		}
		defer f.Close()
		return b.collection.Encode(f)

	})

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

	// write down at which document we arrived
	{
		f, err := os.OpenFile(filepath.Join(path, "stats.bin"), os.O_RDWR|os.O_CREATE, 0600)
		if err != nil {
			return err
		}

		s, _ := stats.Load(f)
		s.IndexingSettings = b.IndexingSettings
		s.Update(uint32(b.collection.Len()), b.collection.AvgDocumentSize)
		if err := s.Dump(f); err != nil {
			return err
		}

		defer f.Close()
	}

	b.lexicon.Clear()
	b.collection.Clear()
	return nil
}

func ReadPartitions(path string) ([]fs.DirEntry, error) {
	partitions, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	return slices.DeleteFunc(partitions, func(p fs.DirEntry) bool { return !p.IsDir() }), nil
}
