package ddl

import (
	"errors"
)

var (
	ErrUnexpectedCurrentToken  = errors.New("unexpected current token")
	ErrUnexpectedPeekToken     = errors.New("unexpected peek token")
	ErrUnexpectedToken         = errors.New("unexpected token")
	ErrNoDifference            = errors.New("no difference")
	ErrNotSupported            = errors.New("not supported")
	ErrAlterOptionNotSupported = errors.New("alter option not supported")
)
