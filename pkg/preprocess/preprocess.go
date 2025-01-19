package preprocess

import (
	"strings"

	"github.com/just-hms/pulse/pkg/word"
)

type Settings struct {
	StopWordsRemoval bool
	Stemming         bool
}

// Tokens returns the tokens of a given document following this steps:
//
//   - to lowercase
//   - clean up (remove control/malformed unicode characters and punctioation)
//   - tokenization
//   - <stopwords removal>
//   - <stemming>
//
// the last two step are optional
func Tokens(content string, s Settings) []string {
	content = strings.ToLower(content)
	content = word.Clean(content)
	tokens := word.Tokenize(content)
	if s.StopWordsRemoval {
		tokens = word.StopWordsRemoval(tokens)
	}
	if s.Stemming {
		word.Stem(tokens)
	}
	return tokens
}

func Frequencies(tokens []string) map[string]uint32 {
	freqs := make(map[string]uint32, len(tokens)/2)
	for _, term := range tokens {
		if _, ok := freqs[term]; !ok {
			freqs[term] = 1
		} else {
			freqs[term]++
		}
	}
	return freqs
}
