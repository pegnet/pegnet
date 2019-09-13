package common

import (
	"context"

	log "github.com/sirupsen/logrus"
)

var GlobalExitHandler = NewExitHandler()

type ExitHandler struct {
	ClosingFunctions []func() error
}

func NewExitHandler() *ExitHandler {
	e := new(ExitHandler)

	return e
}

func (e *ExitHandler) AddExit(f func() error) {
	e.ClosingFunctions = append(e.ClosingFunctions, f)
}

func (e *ExitHandler) AddCancel(cancel context.CancelFunc) {
	e.ClosingFunctions = append(e.ClosingFunctions, func() error {
		cancel()
		return nil
	})
}

func (e *ExitHandler) Close() {
	for _, f := range e.ClosingFunctions {
		err := f()
		if err != nil {
			log.WithError(err).Errorf("failed to close")
		}
	}
}
