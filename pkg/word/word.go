package word

import (
	"regexp"
	"slices"
	"strings"

	_ "embed"

	"github.com/blevesearch/go-porterstemmer"
	"golang.org/x/sync/errgroup"
	"golang.org/x/text/transform"
)

var (
	punctuationRemover = regexp.MustCompile(`[^\p{L}\p{N}\p{So}\s]+`)
	htmlTags           = regexp.MustCompile(`<[^>]*>`)
)

func Clean(s string) string {
	s, _, _ = transform.String(unicodeNormalizer(), s)
	s = htmlTags.ReplaceAllString(s, " ")
	s = punctuationRemover.ReplaceAllString(s, " ")
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
	wg := errgroup.Group{}
	for i, t := range tokens {
		wg.Go(func() error {
			tokens[i] = string(
				porterstemmer.StemWithoutLowerCasing([]rune(t)),
			)
			return nil
		})
	}
	wg.Wait()
}
