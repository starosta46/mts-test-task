package validator

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/mts-test-task/pkg/sitesdataservice/httperror"
)

// Input validate input data
type Input interface {
	CheckURLs(urls []string) (err error)
}

type input struct {
	maxURLsCount int
	errorCreator httperror.ErrorCreator
}

func (i *input) CheckURLs(urls []string) (err error) {
	if len(urls) == 0 {
		return i.errorCreator(
			http.StatusBadRequest,
			"Ошибка ввода: нет ни одного урла",
			fmt.Sprintf("input validation error: %s", "no urls"),
		)
	}

	if len(urls) > i.maxURLsCount {
		return i.errorCreator(
			http.StatusBadRequest,
			"Ошибка ввода: слишком много урлов",
			fmt.Sprintf("input validation error: %s", "too many urls"),
		)
	}

	for j := 0; j < len(urls); j++ {
		_, err = url.ParseRequestURI(urls[j])
		if err != nil {
			return i.errorCreator(
				http.StatusBadRequest,
				fmt.Sprintf("Ошибка ввода: неверный урл: %s", urls[j]),
				fmt.Sprintf("input validation error: %s %s", "bad url:", urls[j]),
			)
		}
	}

	return
}

// NewInput ...
func NewInput(maxURLsCount int, errorCreator httperror.ErrorCreator) Input {
	return &input{
		maxURLsCount: maxURLsCount,
		errorCreator: errorCreator,
	}
}
