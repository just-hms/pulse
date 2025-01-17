package countwriter_test

import (
	"os"
	"testing"

	"github.com/just-hms/pulse/pkg/countwriter"
	"github.com/stretchr/testify/require"
)

func TestWriter(t *testing.T) {
	t.Parallel()

	t.Run("Write updates Count correctly", func(t *testing.T) {
		t.Parallel()
		req := require.New(t)

		// Create a temporary file for testing
		tempFile, err := os.CreateTemp("", "countwriter_test")
		req.NoError(err)
		defer os.Remove(tempFile.Name()) // Clean up the file afterward

		// Create a new Writer instance
		cw := countwriter.NewWriter(tempFile)

		// Write some data
		data := []byte("hello")
		n, err := cw.Write(data)
		req.NoError(err)
		req.Equal(len(data), n, "Write should return the number of bytes written")
		req.Equal(len(data), cw.Count, "Count should match the number of bytes written")

		// Write more data
		moreData := []byte(" world")
		n, err = cw.Write(moreData)
		req.NoError(err)
		req.Equal(len(moreData), n, "Write should return the number of bytes written")
		req.Equal(len(data)+len(moreData), cw.Count, "Count should accumulate the number of bytes written")
	})

	t.Run("Write appends to the file", func(t *testing.T) {
		t.Parallel()
		req := require.New(t)

		// Create a temporary file for testing
		tempFile, err := os.CreateTemp("", "countwriter_test")
		req.NoError(err)
		defer os.Remove(tempFile.Name()) // Clean up the file afterward

		// Create a new Writer instance
		cw := countwriter.NewWriter(tempFile)

		// Write some data
		data := []byte("initial")
		_, err = cw.Write(data)
		req.NoError(err)

		// Write more data
		moreData := []byte(" data")
		_, err = cw.Write(moreData)
		req.NoError(err)

		// Check the file content
		err = tempFile.Close()
		req.NoError(err)

		contents, err := os.ReadFile(tempFile.Name())
		req.NoError(err)
		req.Equal("initial data", string(contents), "File contents should match the written data")
	})
}
