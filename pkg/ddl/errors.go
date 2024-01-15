package ddl

import (
	"errors"
)

var (
	ErrUnexpectedToken         = errors.New("unexpected token")
	ErrNoDifference            = errors.New("no difference")
	ErrNotSupported            = errors.New("not supported")
	ErrAlterOptionNotSupported = errors.New("alter option not supported")
)
