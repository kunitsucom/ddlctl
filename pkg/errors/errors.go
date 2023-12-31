package errors

import "errors"

var (
	ErrNotSupported                       = errors.New("not supported")
	ErrUserCanceled                       = errors.New("user canceled")
	ErrDialectIsEmpty                     = errors.New("dialect is empty")
	ErrDDLTagGoAnnotationNotFoundInSource = errors.New("ddl-tag-go annotation not found in source")
	ErrTwoArgumentsRequired               = errors.New("two arguments required")
	ErrBothArgumentsIsDSN                 = errors.New("both arguments is dsn")
	ErrBothArgumentsAreNotDSNOrSQLFile    = errors.New("both arguments are not dsn or sql file")
)
