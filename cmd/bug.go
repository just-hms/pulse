package main

import (
	"fmt"

	"github.com/blevesearch/go-porterstemmer"
)

func main() {
	start := []rune("eing")
	end := porterstemmer.Stem(start)
	fmt.Println(string(end))
}
