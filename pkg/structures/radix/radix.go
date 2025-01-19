package radix

import (
	"encoding/gob"
	"errors"
	"io"
	"iter"

	iradix "github.com/hashicorp/go-immutable-radix/v2"
	"github.com/just-hms/pulse/pkg/structures/withkey"
)

// Decode deocdes iradix.Tree from a io.Reader
func Decode[T any](r io.Reader, tree **iradix.Tree[T]) error {
	dec := gob.NewDecoder(r)

	txn := (*tree).Txn()
	for {
		t := withkey.WithKey[T]{}
		err := dec.Decode(&t)
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return err
		}

		txn.Insert([]byte(t.Key), t.Value)
	}

	(*tree) = txn.Commit()
	return nil
}

// Values returns an iterator over the iradix.Tree values
func Values[T any](tree *iradix.Tree[T]) iter.Seq2[[]byte, T] {
	it := tree.Root().Iterator()
	return func(yield func([]byte, T) bool) {
		for key, val, ok := it.Next(); ok; key, val, ok = it.Next() {
			if !yield(key, val) {
				return
			}
		}
	}
}

// Encode encodes a iradix.Tree into a io.Writer
func Encode[T any](w io.Writer, tree *iradix.Tree[T]) error {
	enc := gob.NewEncoder(w)

	for key, val := range Values(tree) {
		t := withkey.WithKey[T]{
			Key:   string(key),
			Value: val,
		}

		if err := enc.Encode(t); err != nil {
			return err
		}
	}
	return nil
}
