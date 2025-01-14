package preprocess

import (
	"strings"

	"github.com/just-hms/pulse/pkg/word"
)

func GetTokens(content string) []string {
	content = strings.ToLower(content)
	content = word.Clean(content)
	tokens := word.Tokenize(content)
	tokens = word.StopWordsRemoval(tokens)
	word.Stem(tokens)
	return tokens
}
