package reader_test

import (
	"bufio"
	"os"
	"testing"

	"github.com/just-hms/pulse/pkg/spimi"
	"github.com/just-hms/pulse/pkg/spimi/reader"
	"github.com/stretchr/testify/require"
)

func TestMsMarco(t *testing.T) {
	t.Parallel()
	req := require.New(t)

	f, err := os.Open("testdata/test.tsv")
	req.NoError(err)

	r := reader.NewMsMarco(bufio.NewReader(f), 5)

	got := [][]spimi.Document{}
	for v, err := range r.Read() {
		req.NoError(err)
		got = append(got, v)
	}

	req.Len(got, 3)
	req.Len(got[0], 5)
	req.Len(got[1], 5)
	req.Len(got[2], 2)
}
