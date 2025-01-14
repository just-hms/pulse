package word

import (
	"regexp"
	"slices"
	"strings"
	"sync"

	_ "embed"

	"github.com/blevesearch/go-porterstemmer"
	"golang.org/x/text/transform"
)

var (
	punctuationRemover = regexp.MustCompile(`[^\w\s]+`) // Match punctuation (non-alphanumeric, non-whitespace)
	htmlTags           = regexp.MustCompile(`<[^>]*>`)  // Match HTML tags
)

func Clean(s string) string {
	s, _, _ = transform.String(unicodeNormalizer(), s)
	s = htmlTags.ReplaceAllString(s, "")
	s = punctuationRemover.ReplaceAllString(s, "")
	return s
}

func Tokenize(s string) []string {
	return strings.Fields(s)
}

func StopWordsRemoval(tokens []string) []string {
	return slices.DeleteFunc(tokens, func(token string) bool {
		return stopWords.Has(token)
	})
}

func Stem(tokens []string) {
	wg := sync.WaitGroup{}
	wg.Add(len(tokens))
	for i, t := range tokens {
		go func() {
			defer wg.Done()
			tokens[i] = string(
				porterstemmer.StemWithoutLowerCasing([]rune(t)),
			)
		}()
	}

	wg.Wait()
}
