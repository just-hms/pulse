package inverseindex

type GlobalTerm struct {
	Frequence         uint32 // number of times a term appears in the collection
	DocumentFrequency uint32 // number of documents in which the term appears
	MaxDocFrequence   uint32 // max frequence inside a document
}

type LocalTerm struct {
	GlobalTerm
	StartOffset uint32
	EndOffset   uint32
}
