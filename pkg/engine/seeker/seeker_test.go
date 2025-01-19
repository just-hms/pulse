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

		lex.Add(map[string]uint32{"a": 1, "b": 2}, 0)
		lex.Add(map[string]uint32{"b": 3, "c": 4}, 1)
		lex.Add(map[string]uint32{"a": 5, "e": 6}, 2)
		lex.Add(map[string]uint32{"b": 7, "d": 8}, 3)

		f, err := inverseindex.CreateLexicon(dir)
		req.NoError(err)
		defer f.Close()

		err = lex.Encode(f, false)
		req.NoError(err)
	}

	{
		r, err := inverseindex.OpenLexicon(dir)
		req.NoError(err)
		defer r.Close()

		lLexicon := iradix.New[uint32]()
		err = radix.Decode(r.Terms, &lLexicon)
		req.NoError(err)

		testcases := []struct {
			key string
			exp map[uint32]uint32
		}{
			{
				key: "a", exp: map[uint32]uint32{0: 1, 2: 5},
			},
			{
				key: "b", exp: map[uint32]uint32{0: 2, 1: 3, 3: 7},
			},
			{
				key: "c", exp: map[uint32]uint32{1: 4},
			},
			{
				key: "d", exp: map[uint32]uint32{3: 8},
			},
			{
				key: "e", exp: map[uint32]uint32{2: 6},
			},
		}

		for _, tt := range testcases {
			offset, ok := lLexicon.Get([]byte(tt.key))
			req.True(ok)

			term, err := inverseindex.DecodeTerm[inverseindex.LocalTerm](offset, r.TermsInfo)
			req.NoError(err)

			s := seeker.NewSeeker(
				r.Posting, r.Freqs,
				withkey.WithKey[inverseindex.LocalTerm]{Key: tt.key, Value: *term},
				false,
			)

			got := map[uint32]uint32{}

			for !seeker.EOD(s) {
				err := s.Next()
				req.NoError(err)
				got[s.DocumentID] = s.TermFrequency
			}

			req.Equal(tt.exp, got, tt.key)
		}
	}

}
