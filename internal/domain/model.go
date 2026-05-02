package domain

type QuizSession struct {
	SessionItems []SessionItem
}

type SessionItem struct {
	CollectionItem CollectionItem
	ActualAnswer   string
}

type CollectionItem struct {
	Question string
	Answer   string
}

type RangeQuery struct {
	Parts []RangeQueryPart
}

type RangeQueryPart struct {
	OpenIndex  int
	CloseIndex int
}
