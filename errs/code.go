package errs

import (
	"fmt"
	"strconv"
)

type Code uint32

const (
	// Internal is the generic error that maps to HTTP 500.
	Internal Code = iota + 100001
	// NotFound indicates a given resource is not found.
	NotFound
	// Forbidden indicates the user doesn't have the permission to
	// perform given operation.
	Forbidden
	// Unauthenticated indicates the oauth2 authentication failed.
	Unauthenticated
	// InvalidArgument indicates the input is invalid.
	InvalidArgument
	// InvalidConfig indicates the config is invalid.
	InvalidConfig
	// Conflict indicates a database transactional conflict happens.
	Conflict
	// TryAgain indicates a temporary outage and retry
	// could eventually lead to success.
	TryAgain
)

func (c Code) String() string {
	switch c {
	case Internal:
		return "Internal"
	case NotFound:
		return "NotFound"
	case Forbidden:
		return "Forbidden"
	case Unauthenticated:
		return "Unauthenticated"
	case InvalidArgument:
		return "InvalidArgument"
	case InvalidConfig:
		return "InvalidConfig"
	case Conflict:
		return "Conflict"
	case TryAgain:
		return "TryAgain"
	default:
		return "Code(" + strconv.FormatInt(int64(c), 10) + ")"
	}
}

func (c Code) Wrap(err error) error {
	_, ok := err.(*Error)
	if ok {
		return err
	}

	return &Error{
		Code: c,
		Msg:  err.Error(),
	}
}

func (c Code) New(a ...string) error {
	msg := ""
	for i, s := range a {
		if i > 0 {
			msg += " "
		}
		msg += s
	}
	return &Error{
		Code: c,
		Msg:  msg,
	}
}

func (c Code) Newf(msg string, args ...interface{}) error {
	return &Error{
		Code: c,
		Msg:  fmt.Sprintf(msg, args...),
	}
}

func (c Code) Is(err error) bool {
	v, ok := err.(*Error)
	if !ok {
		// all other errors are internal error.
		return c == Internal
	}
	return v.Code == c
}
