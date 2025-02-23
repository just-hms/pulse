package word

import (
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

var (
	controlCharacters = runes.Predicate(func(r rune) bool {
		return unicode.Is(unicode.C, r) && r != '\n' && r != '\t'
	})
)

// unicodeNormalizer removes malformed unicode and control character beside '\n' and `\t`
func unicodeNormalizer() transform.Transformer {
	return transform.Chain(
		norm.NFD,                           // Decompose Unicode
		runes.Remove(runes.In(unicode.Mn)), // Remove diacritics (non-spacing marks)
		runes.Remove(controlCharacters),    // Remove control characters
		norm.NFC,                           // Recompose Unicode
	)
}
