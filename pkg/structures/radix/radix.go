package radix

import (
	"encoding/gob"
	"errors"
	"fmt"
	"io"
	"iter"
	"log"

	iradix "github.com/hashicorp/go-immutable-radix/v2"
	"github.com/just-hms/pulse/pkg/structures/withkey"
)

func Decode[T any](r io.Reader, tree **iradix.Tree[T]) error {
	// todo: remove
	log.Println("checkpoint 1")
	fmt.Scanln()

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

		log.Println(t.Key, t.Value)
		fmt.Scanln()

		*tree, _, _ = (*tree).Insert([]byte(t.Key), t.Value)
	}

	// todo: remove
	log.Println("checkpoint 2 ")
	fmt.Scanln()

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

		if err := enc.Encode(t); err != nil {
			return err
		}
	}
	return nil
}
