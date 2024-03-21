package inverseindex

type Term struct {
	Value       string
	DocFreq     uint32
	StartOffset uint32
	EndOffset   uint32
}
