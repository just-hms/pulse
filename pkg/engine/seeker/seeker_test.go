package seeker_test

import (
	"log"
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

		log.Println("aaaaa", aTerm.StartOffset, aTerm.EndOffset)

		s := seeker.NewSeeker(
			f.Posting, f.Freqs,
			withkey.WithKey[inverseindex.LocalTerm]{Key: key, Value: aTerm},
		)

		log.Println("aaaaa", s.Term.Value.StartOffset, s.Term.Value.EndOffset)

		got := map[uint32]uint32{}

		err = s.Next()
		req.NoError(err)
		got[s.DocumentID] = s.Frequence
		err = s.Next()
		req.NoError(err)
		got[s.DocumentID] = s.Frequence

		req.True(seeker.EOD(s))
		err = s.Next()
		req.Error(err)

		exp := map[uint32]uint32{
			0: 2,
			2: 16,
		}
		req.Equal(exp, got)
	}

}
