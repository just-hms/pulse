package inverseindex_test

import (
	"os"
	"testing"

	"github.com/just-hms/pulse/pkg/spimi/inverseindex"
	"github.com/stretchr/testify/require"
)

func TestDocumentIO(t *testing.T) {
	t.Parallel()
	req := require.New(t)

	f, err := os.CreateTemp("", "doc*")
	req.NoError(err)

	defer f.Close()

	col := inverseindex.Collection{
		inverseindex.NewDocument("1", 10),
		inverseindex.NewDocument("2", 2),
		inverseindex.NewDocument("4", 13),
		inverseindex.NewDocument("19", 24),
	}

	err = col.Encode(f)
	req.NoError(err)

	doc := inverseindex.Document{}
	err = doc.Decode(2, f)
	req.NoError(err)

	req.Equal([8]byte{'4'}, doc.No)
	req.Equal(uint32(13), doc.Size)
}
