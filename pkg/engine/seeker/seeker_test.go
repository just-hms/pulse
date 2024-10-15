package seeker_test

import (
	"os"
	"os/exec"
	"testing"

	iradix "github.com/hashicorp/go-immutable-radix/v2"
	"github.com/just-hms/pulse/pkg/engine/seeker"
	"github.com/just-hms/pulse/pkg/spimi/inverseindex"
	"github.com/just-hms/pulse/pkg/structures/radix"
	"github.com/just-hms/pulse/pkg/structures/withkey"
	"github.com/stretchr/testify/require"
)

func Code(f *os.File) {
	c := exec.Command("code", f.Name())
	c.Run()
}

func TestSeeker(t *testing.T) {
	t.Parallel()
	req := require.New(t)

	dir := t.TempDir()
	dir = "/tmp/"

	{
		lex := inverseindex.Lexicon{}

		lex.Add(map[string]uint32{"a": 2, "b": 3}, 0)
		lex.Add(map[string]uint32{"c": 2, "b": 3}, 1)
		lex.Add(map[string]uint32{"a": 16, "e": 3}, 2)
		lex.Add(map[string]uint32{"d": 2, "b": 3}, 3)

		f, err := inverseindex.CreateLexiconFiles(dir)
		req.NoError(err)
		defer f.Close()

		err = lex.Encode(f)
		req.NoError(err)
	}
	{
		f, err := inverseindex.OpenLexiconFiles(dir)
		req.NoError(err)
		defer f.Close()

		localLexicon := iradix.New[inverseindex.LocalTerm]()
		err = radix.Decode(f.TermsFile, &localLexicon)
		req.NoError(err)

		key := "a"
		aTerm, ok := localLexicon.Get([]byte(key))
		req.True(ok)

		seeker := seeker.NewSeeker(
			f.PostingFile, f.FreqsFile,
			withkey.WithKey[inverseindex.LocalTerm]{Key: key, Value: aTerm},
		)

		seeker.Next()
		req.Equal(uint32(2), seeker.Frequence)
		req.Equal(uint32(0), seeker.DocumentID)

		seeker.Next()
		req.Equal(uint32(16), seeker.Frequence)
		req.Equal(uint32(2), seeker.DocumentID)
	}

}
