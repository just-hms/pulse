package seeker_test

import (
	"testing"

	iradix "github.com/hashicorp/go-immutable-radix/v2"
	"github.com/just-hms/pulse/pkg/engine/seeker"
	"github.com/just-hms/pulse/pkg/spimi/inverseindex"
	"github.com/just-hms/pulse/pkg/structures/radix"
	"github.com/just-hms/pulse/pkg/structures/withkey"
	"github.com/stretchr/testify/require"
)

func TestSeeker(t *testing.T) {
	t.Parallel()
	req := require.New(t)

	dir := t.TempDir()
	{
		lex := inverseindex.Lexicon{}

		lex.Add(map[string]uint32{"a": 2, "b": 3}, 0)
		lex.Add(map[string]uint32{"c": 2, "b": 3}, 1)
		lex.Add(map[string]uint32{"a": 16, "e": 3}, 2)
		lex.Add(map[string]uint32{"d": 2, "b": 3}, 3)

		f, err := inverseindex.CreateLexicon(dir)
		req.NoError(err)
		defer f.Close()

		err = lex.Encode(f)
		req.NoError(err)
	}
	{
		f, err := inverseindex.OpenLexicon(dir)
		req.NoError(err)
		defer f.Close()

		localLexicon := iradix.New[inverseindex.LocalTerm]()
		err = radix.Decode(f.Terms, &localLexicon)
		req.NoError(err)

		key := "a"
		aTerm, ok := localLexicon.Get([]byte(key))
		req.True(ok)

		seeker := seeker.NewSeeker(
			f.Posting, f.Freqs,
			withkey.WithKey[inverseindex.LocalTerm]{Key: key, Value: aTerm},
		)

		err = seeker.Next()
		req.NoError(err)

		req.Equal(uint32(0), seeker.DocumentID)
		req.Equal(uint32(2), seeker.Frequence)

		err = seeker.Next()
		req.NoError(err)
		req.Equal(uint32(2), seeker.DocumentID)
		req.Equal(uint32(16), seeker.Frequence)
	}

}
