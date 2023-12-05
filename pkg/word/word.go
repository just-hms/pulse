package word

import (
	"slices"
	"strings"
	"sync"

	"github.com/blevesearch/go-porterstemmer"
)

var punctuationRemover = strings.NewReplacer(
	".", "", ",", "", "!", "", "?", "",
	";", "", ":", "", "'", "", "\"", "",
	"(", "", ")", "", "[", "", "]", "",
	"{", "", "}", "",
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
		"home", "can", "us", "about",
		"if", "page", "my", "has",
		"search", "free", "but", "our",
		"one", "other", "do", "no",
		"information", "time", "they", "site",
		"he", "up", "may", "what",
		"which", "their", "news", "out",
		"use", "any", "see", "only",
		"so", "his", "when", "contact",
		"here", "business", "who", "web",
		"also", "now", "help", "get",
		"pm", "view", "online", "first",
		"am", "been", "would", "how",
		"were", "me", "services", "some",
		"these", "click", "like", "service",
		"x", "than", "find", "price",
		"date", "back", "top", "people",
		"had", "list", "name", "just",
		"over", "state", "year", "day",
		"into", "email", "two", "health",
		"n", "world", "re", "next",
		"used", "go", "b", "work",
		"last", "most", "music", "buy",
		"data", "make", "them", "should",
		"product", "system", "post", "her",
		"city", "t", "add", "policy",
		"number", "such", "please", "great",
		"united", "hotel", "real", "f",
		"item", "international", "center", "ebay",
		"must", "store", "travel", "comments",
		"made", "report", "off", "member",
		"details", "line", "terms", "before",
		"hotels", "did", "send", "right",
		"type", "because", "local", "those", "eing", "":
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
			kek := []rune(t)
			tokens[i] = string(porterstemmer.StemWithoutLowerCasing(kek))
		}(i, t)
	}

	wg.Wait()
}
