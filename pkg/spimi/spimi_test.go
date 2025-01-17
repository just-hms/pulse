package spimi_test

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"maps"
	"os"
	"path/filepath"
	"testing"

	iradix "github.com/hashicorp/go-immutable-radix/v2"
	"github.com/just-hms/pulse/pkg/spimi"
	"github.com/just-hms/pulse/pkg/spimi/inverseindex"
	"github.com/just-hms/pulse/pkg/spimi/readers"
	"github.com/just-hms/pulse/pkg/spimi/stats"
	"github.com/just-hms/pulse/pkg/structures/radix"
	"github.com/stretchr/testify/require"
)

func TestSpimi(t *testing.T) {
	t.Parallel()
	req := require.New(t)

	s := stats.IndexingSettings{
		Stemming:          false,
		StopWordsRemoval:  false,
		Compression:       false,
		MemoryThresholdMB: spimi.DefaulMemoryThreshold,
	}

	spimiBuilder := spimi.NewBuilder(s)
	path := "/tmp/pulse/dump"

	for i := range 2 {
		dataset := fmt.Sprintf("testdata/little%d.tsv", i)

		f, err := os.Open(dataset)
		req.NoError(err)

		r := readers.NewMsMarco(bufio.NewReader(f), 10)

		err = spimiBuilder.Parse(r, 1, path)
		req.NoError(err)
	}
	err := spimi.Merge(path)
	req.NoError(err)

	gTermsF, err := os.Open(filepath.Join(path, "terms.bin"))
	req.NoError(err)

	gTermInfoF, err := os.Open(filepath.Join(path, "terms-info.bin"))
	req.NoError(err)

	gLexicon := iradix.New[uint32]()
	err = radix.Decode(gTermsF, &gLexicon)
	req.NoError(err)

	exp := map[string]struct {
		Postings          [][]uint32
		Frequencies       [][]uint32
		MaxTermFrequency  uint32
		DocumentFrequency uint32
	}{
		"the": {
			[][]uint32{{0, 1, 2}, {3, 5}},
			[][]uint32{{3, 1, 2}, {2, 2}},
			3,
			5,
		},
	}

	tokens := maps.Keys(exp)

	for token := range tokens {
		expInfo := exp[token]

		gOffset, ok := gLexicon.Get([]byte(token))
		req.True(ok)

		gTerm, err := inverseindex.DecodeTerm[inverseindex.GlobalTerm](gOffset, gTermInfoF)
		req.NoError(err)

		req.Equal(expInfo.DocumentFrequency, gTerm.DocumentFrequency)
		req.Equal(expInfo.MaxTermFrequency, gTerm.MaxTermFrequency)

		for i := range 2 {
			readers, err := spimi.OpenSpimiFiles(filepath.Join(path, fmt.Sprint(i)))
			req.NoError(err)

			lLexicon := iradix.New[uint32]()
			err = radix.Decode(readers.Terms, &lLexicon)
			req.NoError(err)

			lOffset, ok := lLexicon.Get([]byte(token))
			req.True(ok)

			lTerm, err := inverseindex.DecodeTerm[inverseindex.LocalTerm](lOffset, readers.TermsInfo)
			req.NoError(err)

			// check postings
			{
				content, err := io.ReadAll(
					io.NewSectionReader(readers.Posting, int64(lTerm.PostStart), int64(lTerm.PostLength)),
				)
				req.NoError(err)

				var buf bytes.Buffer
				binary.Write(&buf, binary.LittleEndian, expInfo.Postings[i])

				req.Equal(buf.Bytes(), content)
			}

			// check frequencies
			{
				content, err := io.ReadAll(
					io.NewSectionReader(readers.Freqs, int64(lTerm.FreqStart), int64(lTerm.FreqLength)),
				)
				req.NoError(err)

				var buf bytes.Buffer
				binary.Write(&buf, binary.LittleEndian, expInfo.Frequencies[i])

				req.Equal(buf.Bytes(), content)
			}

		}
	}

}
