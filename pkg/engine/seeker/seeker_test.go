package seeker_test

import (
	"testing"

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

		cf, err := inverseindex.CreateLexiconFiles(dir)
		req.NoError(err)
		defer cf.Close()

		err = lex.Encode(cf)
		req.NoError(err)
	}
	{

		rf, err := inverseindex.OpenLexiconFiles(dir)
		req.NoError(err)
		defer rf.Close()

		localLexicon := radix.New[inverseindex.LocalTerm]()
		err = localLexicon.Decode(rf.TermsFile)
		req.NoError(err)

		key := "a"
		aTerm, ok := localLexicon.Get([]byte(key))
		req.True(ok)

		seeker := seeker.NewSeeker(
			rf.PostingFile, rf.FreqsFile,
			withkey.WithKey[inverseindex.LocalTerm]{Key: key, Value: *aTerm},
		)

		seeker.Next()
		req.Equal(uint32(2), seeker.Frequence)
		req.Equal(uint32(0), seeker.ID)

		seeker.Next()
		req.Equal(uint32(16), seeker.Frequence)
		req.Equal(uint32(2), seeker.ID)
	}

}
