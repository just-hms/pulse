package word

import (
	"slices"
	"strings"
	"sync"

	"github.com/blevesearch/go-porterstemmer"
)

var punctuationRemover = strings.NewReplacer(
	".", " ", ",", " ", "!", " ", "?", " ",
	";", " ", ":", " ", "'", " ", "\" ", " ",
	"(", " ", ")", " ", "[", " ", "]", " ",
	"{", " ", "}", " ",
)

func Tokenize(s string) []string {
	s = strings.ToLower(s)
	s = punctuationRemover.Replace(s)
	return strings.Fields(s)
}

// isStopword checks if a word is a stopword.
func isStopword(word string) bool {
	switch word {
	case "the", "and", "in", "of",
		"to", "a", "for", "is",
		"on", "that", "by", "this",
		"with", "i", "you", "it",
		"not", "or", "be", "are",
		"from", "at", "as", "your",
		"all", "have", "new", "more",
		"an", "was", "we", "will",
		"can", "if", "my", "has",
		"but", "our", "one", "do", "no",
		"he", "up", "may", "out",
		"use", "any", "see", "only",
		"so", "his", "when", "here", "who",
		"now", "get", "pm", "view", "am",
		"been", "how", "were", "me", "some":
		return true
	}
	return false
}

func StopWordsRemoval(tokens []string) []string {
	return slices.DeleteFunc(tokens, func(el string) bool {
		return isStopword(el)
	})
}

func Stem(tokens []string) {
	wg := sync.WaitGroup{}
	wg.Add(len(tokens))
	for i, t := range tokens {
		go func(i int, t string) {
			defer wg.Done()
			tokens[i] = string(
				porterstemmer.StemWithoutLowerCasing([]rune(t)),
			)
		}(i, t)
	}

	wg.Wait()
}
