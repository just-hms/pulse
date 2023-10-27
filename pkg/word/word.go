package word

import (
	"slices"
	"strings"

	"github.com/kljensen/snowball/english"
)

var punctuationRemover = strings.NewReplacer(
	".", "", ",", "", "!", "", "?", "", ":", "", ";", "",
	"(", "", ")", "", "[", "", "]", "", "{", "", "}", "",
)

func Tokenize(s string) []string {
	s = punctuationRemover.Replace(s)
	return strings.Fields(s)
}

var full = struct{}{}
var stopwords = map[string]struct{}{
	"the": full,
	"and": full,
	"in":  full,
	"of":  full,
	"to":  full,
	"a":   full,
}

func StopWordsRemoval(tokens []string) {
	slices.DeleteFunc(tokens, func(el string) bool {
		_, ok := stopwords[el]
		return ok
	})
}

func Stem(tokens []string) {
	for i, t := range tokens {
		tokens[i] = english.Stem(t, true)
	}
}
