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

	// create a collection
	col := inverseindex.Collection{}
	// add some documents
	col.Add(
		inverseindex.NewDocument("1", 10),
		inverseindex.NewDocument("2", 2),
		inverseindex.NewDocument("4", 13),
		inverseindex.NewDocument("19", 24),
	)

	// encode the collection into a file
	err = col.Encode(f)
	req.NoError(err)

	// check that the decode function works
	doc := inverseindex.Document{}
	{
		err := doc.Decode(2, f)
		req.NoError(err)

		req.Equal("4", doc.No())
		req.Equal(uint32(13), doc.Size)
	}
	{
		err := doc.Decode(3, f)
		req.NoError(err)

		req.Equal("19", doc.No())
		req.Equal(uint32(24), doc.Size)
	}
}
