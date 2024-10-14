package radix

import (
	"encoding/gob"
	"errors"
	"io"

	iradix "github.com/hashicorp/go-immutable-radix/v2"
	"github.com/just-hms/pulse/pkg/structures/withkey"
)

type Tree[T any] struct {
	*iradix.Tree[*T]
}

func New[T any]() *Tree[T] {
	return &Tree[T]{
		Tree: iradix.New[*T](),
	}
}

func (l *Tree[T]) Decode(r io.Reader) error {
	txn := l.Tree.Txn()
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
		txn.Insert([]byte(t.Key), &t.Value)
	}

	l.Tree = txn.Commit()
	return nil
}

// todo: check pointers

func (l *Tree[T]) Append(other *Tree[T], merger func(a, b T) T) {
	txn := l.Tree.Txn()
	it := other.Tree.Root().Iterator()
	for key, t, ok := it.Next(); ok; key, _, ok = it.Next() {
		if v, ok := txn.Get(key); ok {
			*v = merger(*v, *t)
		} else {
			txn.Insert(key, t)
		}
	}
	l.Tree = txn.Commit()
}

func (l *Tree[T]) Encode(w io.Writer) error {
	enc := gob.NewEncoder(w)

	it := l.Tree.Root().Iterator()
	for key, t, ok := it.Next(); ok; key, _, ok = it.Next() {
		t := withkey.WithKey[*T]{
			Key:   string(key),
			Value: t,
		}
		err := enc.Encode(&t)
		if err != nil {
			return err
		}
	}
	return nil
}
