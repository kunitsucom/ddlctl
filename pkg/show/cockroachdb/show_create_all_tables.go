package cockroachdb

import (
	"context"
	"database/sql"

	sqlz "github.com/kunitsucom/util.go/database/sql"
	errorz "github.com/kunitsucom/util.go/errors"
)

type sqlQueryerContext = interface {
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
}

const (
	queryShowCreateAllTables = `-- CREATE TABLE
SHOW CREATE ALL TABLES
;
`
)

func ShowCreateAllTables(ctx context.Context, db sqlQueryerContext) (query string, err error) {
	dbz := sqlz.NewDB(db)

	type CreateStatement struct {
		CreateStatement string `db:"create_statement"`
	}

	createTableStmts := new([]*CreateStatement)
	if err := dbz.QueryContext(ctx, createTableStmts, queryShowCreateAllTables); err != nil {
		return "", errorz.Errorf("dbz.QueryContext: %w", err)
	}
	for _, stmt := range *createTableStmts {
		query += stmt.CreateStatement + "\n"
	}

	return query, nil
}
