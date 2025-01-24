package ranges

type RangeQuery struct {
	Parts []RangeQueryPart
}

type RangeQueryPart struct {
	OpenIndex  int
	CloseIndex int
}
