package preprocess

import (
	"strings"

	"github.com/just-hms/pulse/pkg/word"
	"golang.org/x/text/transform"
)

func GetTokens(content string) []string {
	content, _, _ = transform.String(word.UnicodeNormalizer(), content)
	content = strings.ToLower(content)
	tokens := word.Tokenize(content)
	tokens = word.StopWordsRemoval(tokens)
	word.Stem(tokens)
	return tokens
}
