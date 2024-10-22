package inverseindex

type GlobalTerm struct {
	Frequence       uint32
	Appearences     uint32
	MaxDocFrequence uint32
}

type LocalTerm struct {
	GlobalTerm
	StartOffset uint32
	EndOffset   uint32
}
