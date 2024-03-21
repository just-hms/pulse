package preprocess

import "github.com/just-hms/pulse/pkg/word"

func GetTokens(content string) []string {
	tokens := word.Tokenize(content)
	tokens = word.StopWordsRemoval(tokens)
	word.Stem(tokens)
	return tokens
}
