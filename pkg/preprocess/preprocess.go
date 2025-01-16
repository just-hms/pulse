package preprocess

import (
	"strings"

	"github.com/just-hms/pulse/pkg/spimi/stats"
	"github.com/just-hms/pulse/pkg/word"
)

func GetTokens(content string, s stats.IndexingSettings) []string {
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
