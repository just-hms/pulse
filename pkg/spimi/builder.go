package spimi

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/just-hms/pulse/pkg/spimi/inverseindex"
	"github.com/just-hms/pulse/pkg/spimi/stats"
	"golang.org/x/sync/errgroup"
)

type builder struct {
	Lexicon    inverseindex.Lexicon
	Collection inverseindex.Collection
	mu         sync.Mutex

	dumpCounter uint32
}

func newBuilder() *builder {
	return &builder{
		Lexicon:     inverseindex.Lexicon{},
		Collection:  inverseindex.Collection{},
		mu:          sync.Mutex{},
		dumpCounter: 0,
	}
}

func (b *builder) Add(freqs map[string]uint32, doc inverseindex.Document) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.Collection.Add(doc)
	b.Lexicon.Add(freqs, uint32(b.Collection.Len()-1))
}

func (b *builder) Encode(path string) error {
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
		return b.Collection.Encode(f)

	})

	wg.Go(func() error {
		f, err := inverseindex.CreateLexicon(partitionPath)
		if err != nil {
			return err
		}
		defer f.Close()
		return b.Lexicon.Encode(f)
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
		s.CollectionSize += uint32(b.Collection.Len())

		err = s.Dump(f)
		if err != nil {
			return err
		}

		defer f.Close()
	}

	// make this better, maybe use a double buffer
	b.Lexicon.Clear()
	b.Collection.Clear()
	return nil
}
