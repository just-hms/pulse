package word

import "github.com/just-hms/pulse/pkg/structures/set"

// stopWords is set of stop words
var stopWords = set.Set[string]{
	"a":          struct{}{},
	"about":      struct{}{},
	"above":      struct{}{},
	"after":      struct{}{},
	"again":      struct{}{},
	"against":    struct{}{},
	"all":        struct{}{},
	"am":         struct{}{},
	"an":         struct{}{},
	"and":        struct{}{},
	"any":        struct{}{},
	"are":        struct{}{},
	"aren't":     struct{}{},
	"as":         struct{}{},
	"at":         struct{}{},
	"be":         struct{}{},
	"because":    struct{}{},
	"been":       struct{}{},
	"before":     struct{}{},
	"being":      struct{}{},
	"below":      struct{}{},
	"between":    struct{}{},
	"both":       struct{}{},
	"but":        struct{}{},
	"by":         struct{}{},
	"can't":      struct{}{},
	"cannot":     struct{}{},
	"could":      struct{}{},
	"couldn't":   struct{}{},
	"did":        struct{}{},
	"didn't":     struct{}{},
	"do":         struct{}{},
	"does":       struct{}{},
	"doesn't":    struct{}{},
	"doing":      struct{}{},
	"don't":      struct{}{},
	"down":       struct{}{},
	"during":     struct{}{},
	"each":       struct{}{},
	"few":        struct{}{},
	"for":        struct{}{},
	"from":       struct{}{},
	"further":    struct{}{},
	"had":        struct{}{},
	"hadn't":     struct{}{},
	"has":        struct{}{},
	"hasn't":     struct{}{},
	"have":       struct{}{},
	"haven't":    struct{}{},
	"having":     struct{}{},
	"he":         struct{}{},
	"he'd":       struct{}{},
	"he'll":      struct{}{},
	"he's":       struct{}{},
	"her":        struct{}{},
	"here":       struct{}{},
	"here's":     struct{}{},
	"hers":       struct{}{},
	"herself":    struct{}{},
	"him":        struct{}{},
	"himself":    struct{}{},
	"his":        struct{}{},
	"how":        struct{}{},
	"how's":      struct{}{},
	"i":          struct{}{},
	"i'd":        struct{}{},
	"i'll":       struct{}{},
	"i'm":        struct{}{},
	"i've":       struct{}{},
	"if":         struct{}{},
	"in":         struct{}{},
	"into":       struct{}{},
	"is":         struct{}{},
	"isn't":      struct{}{},
	"it":         struct{}{},
	"it's":       struct{}{},
	"its":        struct{}{},
	"itself":     struct{}{},
	"let's":      struct{}{},
	"me":         struct{}{},
	"more":       struct{}{},
	"most":       struct{}{},
	"mustn't":    struct{}{},
	"my":         struct{}{},
	"myself":     struct{}{},
	"no":         struct{}{},
	"nor":        struct{}{},
	"not":        struct{}{},
	"of":         struct{}{},
	"off":        struct{}{},
	"on":         struct{}{},
	"once":       struct{}{},
	"only":       struct{}{},
	"or":         struct{}{},
	"other":      struct{}{},
	"ought":      struct{}{},
	"our":        struct{}{},
	"ours":       struct{}{},
	"ourselves":  struct{}{},
	"out":        struct{}{},
	"over":       struct{}{},
	"own":        struct{}{},
	"same":       struct{}{},
	"shan't":     struct{}{},
	"she":        struct{}{},
	"she'd":      struct{}{},
	"she'll":     struct{}{},
	"she's":      struct{}{},
	"should":     struct{}{},
	"shouldn't":  struct{}{},
	"so":         struct{}{},
	"some":       struct{}{},
	"such":       struct{}{},
	"than":       struct{}{},
	"that":       struct{}{},
	"that's":     struct{}{},
	"the":        struct{}{},
	"their":      struct{}{},
	"theirs":     struct{}{},
	"them":       struct{}{},
	"themselves": struct{}{},
	"then":       struct{}{},
	"there":      struct{}{},
	"there's":    struct{}{},
	"these":      struct{}{},
	"they":       struct{}{},
	"they'd":     struct{}{},
	"they'll":    struct{}{},
	"they're":    struct{}{},
	"they've":    struct{}{},
	"this":       struct{}{},
	"those":      struct{}{},
	"through":    struct{}{},
	"to":         struct{}{},
	"too":        struct{}{},
	"under":      struct{}{},
	"until":      struct{}{},
	"up":         struct{}{},
	"very":       struct{}{},
	"was":        struct{}{},
	"wasn't":     struct{}{},
	"we":         struct{}{},
	"we'd":       struct{}{},
	"we'll":      struct{}{},
	"we're":      struct{}{},
	"we've":      struct{}{},
	"were":       struct{}{},
	"weren't":    struct{}{},
	"what":       struct{}{},
	"what's":     struct{}{},
	"when":       struct{}{},
	"where":      struct{}{},
	"where's":    struct{}{},
	"which":      struct{}{},
	"while":      struct{}{},
	"who":        struct{}{},
	"who's":      struct{}{},
	"whom":       struct{}{},
	"why":        struct{}{},
	"why's":      struct{}{},
	"with":       struct{}{},
	"won't":      struct{}{},
	"would":      struct{}{},
	"wouldn't":   struct{}{},
	"you":        struct{}{},
	"you'd":      struct{}{},
	"you'll":     struct{}{},
	"you're":     struct{}{},
	"you've":     struct{}{},
	"your":       struct{}{},
	"yours":      struct{}{},
	"yourself":   struct{}{},
	"yourselves": struct{}{},
}
