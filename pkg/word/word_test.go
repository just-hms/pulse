package word_test

import (
	"testing"

	"github.com/just-hms/pulse/pkg/word"
	"github.com/stretchr/testify/require"
)

func TestTokenize(t *testing.T) {
	t.Parallel()
	req := require.New(t)

	tests := []struct {
		name  string
		input string
		exp   []string
	}{
		{
			"Basic",
			"hello, world!",
			[]string{"hello", "world"},
		},
		{
			"Punctuation",
			"it's a test.",
			[]string{"it", "s", "a", "test"},
		},

		{
			"Empty string",
			"",
			[]string{},
		},
		{
			"No spaces",
			"word",
			[]string{"word"},
		},
		{
			"First ms-marco line",
			"the presence of communication amid scientific minds was equally important to the success of the manhattan project as scientific intellect was. the only cloud hanging over the impressive achievement of the atomic researchers and engineers is what their success truly meant; hundreds of thousands of innocent lives obliterated.",
			[]string{
				"the", "presence", "of", "communication", "amid", "scientific", "minds",
				"was", "equally", "important", "to", "the", "success", "of", "the", "manhattan",
				"project", "as", "scientific", "intellect", "was", "the", "only", "cloud",
				"hanging", "over", "the", "impressive", "achievement", "of", "the",
				"atomic", "researchers", "and", "engineers", "is", "what", "their", "success",
				"truly", "meant", "hundreds", "of", "thousands", "of", "innocent", "lives", "obliterated",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cleaned := word.Clean(tt.input)
			got := word.Tokenize(cleaned)
			req.Equal(tt.exp, got)
		})
	}
}

func TestStopWordsRemoval(t *testing.T) {
	t.Parallel()
	req := require.New(t)

	tests := []struct {
		name  string
		input []string
		exp   []string
	}{
		{"Mixed stopwords", []string{"hello", "the", "world"}, []string{"hello", "world"}},
		{"All stopwords", []string{"the", "and", "a"}, []string{}},
		{"No stopwords", []string{"unique", "words"}, []string{"unique", "words"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := word.StopWordsRemoval(tt.input)
			req.Equal(tt.exp, got)
		})
	}
}

func TestStem(t *testing.T) {
	t.Parallel()
	req := require.New(t)

	tests := []struct {
		name  string
		input []string
		exp   []string
	}{
		{"Basic stemming", []string{"running", "jumps"}, []string{"run", "jump"}},
		{"No stemming required", []string{"run", "jump"}, []string{"run", "jump"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := tt.input
			word.Stem(got)
			req.Equal(tt.exp, got)
		})
	}
}
