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
