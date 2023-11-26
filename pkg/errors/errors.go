package errors

import "errors"

var (
	ErrNotSupported                       = errors.New("not supported")
	ErrDialectIsEmpty                     = errors.New("dialect is empty")
	ErrDDLTagGoAnnotationNotFoundInSource = errors.New("ddl-tag-go annotation not found in source")
	ErrDiffRequiresTwoArguments           = errors.New("diff requires two arguments")
	ErrBothArgumentsIsDSN                 = errors.New("both arguments is dsn")
	ErrBothArgumentsAreNotDSNOrSQLFile    = errors.New("both arguments are not dsn or sql file")
)
