package main

import (
	"bufio"
	"os"

	"github.com/just-hms/pulse/pkg/index"
)

func main() {
	f := os.Stdin
	if len(os.Args) > 1 {
		var err error
		f, err = os.Open(os.Args[1])
		if err != nil {
			panic(err)
		}
	}

	reader := bufio.NewReader(f)
	index.Load(reader)

}
