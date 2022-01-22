package task

import (
	"fmt"

	"github.com/pkg/errors"
)

type StatusType int

const (
	Info  StatusType = 0
	Error            = 1
	Fatal            = 2
)

type Status struct {
	Type    StatusType
	Message string
	Error   error
	IsDone  bool
}

func InfoStatus(message string) Status {
	return Status{
		Type:    Info,
		Message: message,
	}
}

func InfoStatusf(fmtString string, args ...any) Status {
	return InfoStatus(fmt.Sprintf(fmtString, args...))
}

func ErrorStatus(err error) Status {
	return Status{
		Type:    Error,
		Error:   err,
		Message: err.Error(),
	}
}

func ErrorStatusWithStack(err error) Status {
	return ErrorStatus(errors.WithStack(err))
}

func ErrorStatusf(fmtString string, args ...any) Status {
	return ErrorStatus(fmt.Errorf(fmtString, args...))
}

func ErrorStatusWrap(err error, message string) Status {
	return ErrorStatus(errors.Wrap(err, message))
}

func ErrorStatusWrapf(err error, fmtString string, args ...any) Status {
	return ErrorStatus(errors.Wrapf(err, fmtString, args...))
}

func ErrorStatusMessagef(err error, fmtString string, args ...any) Status {
	return ErrorStatus(errors.WithMessagef(err, fmtString, args...))
}

func ErrorStatusMessage(err error, message string) Status {
	return ErrorStatus(errors.WithMessage(err, message))
}

func FatalStatus(err error) Status {
	return Status{
		Type:    Fatal,
		Error:   err,
		Message: err.Error(),
	}
}

func FatalStatusWithStack(err error) Status {
	return FatalStatus(errors.WithStack(err))
}

func FatalStatusf(fmtString string, args ...any) Status {
	return FatalStatus(fmt.Errorf(fmtString, args...))
}

func FatalStatusWrapf(err error, fmtString string, args ...any) Status {
	return FatalStatus(errors.Wrapf(err, fmtString, args...))
}

func FatalStatusWrap(err error, message string) Status {
	return FatalStatus(errors.Wrap(err, message))
}

func FatalStatusMessagef(err error, fmtString string, args ...any) Status {
	return FatalStatus(errors.WithMessagef(err, fmtString, args...))
}

func FatalStatusMessage(err error, message string) Status {
	return FatalStatus(errors.WithMessage(err, message))
}
