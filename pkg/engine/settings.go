package engine

// SearchSettings contains the search engine settings
type SearchSettings struct {
	K           int
	Metric      Metric
	Conjunctive bool
}
