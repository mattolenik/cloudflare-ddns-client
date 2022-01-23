package task

import (
	"fmt"
)

type StatusType int

const (
	Info  StatusType = 0
	Error            = 1
	Fatal            = 2
)

type Status[T any] struct {
	Type    StatusType
	Message string
	Error   error
	Token   T
}

type StatusStream[T any] chan Status[T]

func (s *StatusStream[T]) Info(msg string) {
	*s <- Status[T]{
		Type:    Info,
		Message: msg,
	}
}

func (s *StatusStream[T]) Infof(format string, args ...any) {
	*s <- Status[T]{
		Type:    Info,
		Message: fmt.Sprintf(format, args...),
	}
}

func (s *StatusStream[T]) Msg(token T, message string) {
	*s <- Status[T]{
		Type:    Info,
		Message: message,
		Token:   token,
	}
}

func (s *StatusStream[T]) Msgf(token T, format string, args ...any) {
	*s <- Status[T]{
		Type:    Info,
		Message: fmt.Sprintf(format, args...),
		Token:   token,
	}
}

func (s *StatusStream[T]) Error(err error) {
	*s <- Status[T]{
		Type:    Error,
		Error:   err,
		Message: err.Error(),
	}
}

func (s *StatusStream[T]) Errorf(format string, args ...any) {
	err := fmt.Errorf(format, args...)
	*s <- Status[T]{
		Type:    Error,
		Error:   err,
		Message: err.Error(),
	}
}

func (s *StatusStream[T]) Fatal(err error) {
	*s <- Status[T]{
		Type:    Fatal,
		Error:   err,
		Message: err.Error(),
	}
}

func (s *StatusStream[T]) Fatalf(format string, args ...any) {
	err := fmt.Errorf(format, args...)
	*s <- Status[T]{
		Type:    Fatal,
		Error:   err,
		Message: err.Error(),
	}
}
