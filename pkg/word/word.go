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

// Clean cleans up a string by:
//   - normalizing the unicoce characters
//   - removing html tags
//   - removing punctuation
func Clean(s string) string {
	s, _, _ = transform.String(unicodeNormalizer(), s)
	s = htmlTags.ReplaceAllString(s, " ")
	s = punctuationRemover.ReplaceAllString(s, " ")
	return s
}

// Tokenize given a strings returns each tokens separating by white space
//
// " ", "  " and "\t" are treated equally
func Tokenize(s string) []string {
	return strings.Fields(s)
}

// StopWordsRemoval remove the tokens which are contained inside a list of stopwords
func StopWordsRemoval(tokens []string) []string {
	return slices.DeleteFunc(tokens, func(token string) bool {
		return stopWords.Has(token)
	})
}

// Stem uses porterstemmer to the given tokens
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
