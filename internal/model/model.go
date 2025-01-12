package model

type QuizSession struct {
	Items []QuizSessionItem
}

type QuizSessionItem struct {
	Index          int
	Question       string
	ExpectedAnswer string
	ActualAnswer   string
}
