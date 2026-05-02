package run

import (
	"github.com/vlostech/qz/internal/domain"
)

type Args struct {
	RangeString   string
	FilePath      string
	QuestionCount int
}

type sessionState struct {
	quizSession   domain.QuizSession
	questionIdx   int
	chosenIndices map[int]struct{}
}

type stepResult struct {
	nextFunc        nextFunc
	isInputRequired bool
}

type stepArgs struct {
	input string
}

type nextFunc func(state *sessionState, args stepArgs) (stepResult, error)
