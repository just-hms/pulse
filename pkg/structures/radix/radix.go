package radix

import (
	"encoding/gob"
	"errors"
	"fmt"
	"io"
	"iter"

	iradix "github.com/hashicorp/go-immutable-radix/v2"
	"github.com/just-hms/pulse/pkg/structures/withkey"
)

func Decode[T any](r io.Reader, tree **iradix.Tree[T]) error {
	txn := (*tree).Txn()
	dec := gob.NewDecoder(r)
	for {
		t := withkey.WithKey[T]{}
		err := dec.Decode(&t)
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return err
		}
		_, ok := txn.Insert([]byte(t.Key), t.Value)
		if !ok {
			return fmt.Errorf("problems inserting %s", t.Key)
		}
	}

	*tree = txn.Commit()
	return nil
}

func Values[T any](tree *iradix.Tree[T]) iter.Seq2[[]byte, T] {
	it := tree.Root().Iterator()
	return func(yield func([]byte, T) bool) {
		for key, t, ok := it.Next(); ok; key, t, ok = it.Next() {
			if !yield(key, t) {
				return
			}
		}
	}
}

func Encode[T any](w io.Writer, tree *iradix.Tree[T]) error {
	enc := gob.NewEncoder(w)

	for k, t := range Values(tree) {
		t := withkey.WithKey[T]{
			Key:   string(k),
			Value: t,
		}
		err := enc.Encode(t)
		if err != nil {
			return err
		}
	}
	return nil
}
