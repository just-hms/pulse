package radix_test

import (
	"testing"

	iradix "github.com/hashicorp/go-immutable-radix/v2"
	"github.com/just-hms/pulse/pkg/structures/radix"
	"github.com/stretchr/testify/require"
)

func Ref[T any](v T) *T {
	return &v
}

func TestMerge(t *testing.T) {
	t.Parallel()
	req := require.New(t)

	aTree := iradix.New[*int]()
	aTreeTxn := aTree.Txn()
	{
		aTreeTxn.Insert([]byte("a"), Ref(1))
		aTreeTxn.Insert([]byte("b"), Ref(4))
		aTreeTxn.Insert([]byte("c"), Ref(12))
	}

	bTree := iradix.New[*int]()
	{
		txn := bTree.Txn()
		txn.Insert([]byte("d"), Ref(10))
		txn.Insert([]byte("e"), Ref(19))
		txn.Insert([]byte("b"), Ref(4))
		bTree = txn.Commit()
	}

	for bK, bV := range radix.Values(bTree) {
		if aV, ok := aTreeTxn.Get(bK); ok {
			*aV += *bV
		} else {
			aTreeTxn.Insert(bK, bV)
		}
	}

	aTree = aTreeTxn.Commit()

	exp := map[string]int{"a": 1, "b": 8, "c": 12, "d": 10, "e": 19}

	for k, expV := range exp {
		v, ok := aTree.Get([]byte(k))
		req.True(ok)
		req.Equal(expV, *v, "with key %q", k)
	}

}
