package inverseindex

type GlobalTerm struct {
	DocFreq uint32
}

type LocalTerm struct {
	GlobalTerm
	StartOffset uint32
	EndOffset   uint32
}
