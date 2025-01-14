package inverseindex

type GlobalTerm struct {
	DocumentFrequency uint32 // number of documents in which the term appears
	MaxTermFrequency  uint32 // max frequence inside a document
}

type LocalTerm struct {
	GlobalTerm
	StartOffset uint32
	EndOffset   uint32
}
