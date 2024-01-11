package ddlctl

import (
	"context"
	"io"
	"os"

	sqlz "github.com/kunitsucom/util.go/database/sql"
	errorz "github.com/kunitsucom/util.go/errors"

	crdbdump "github.com/kunitsucom/ddlctl/pkg/dump/cockroachdb"
	mydump "github.com/kunitsucom/ddlctl/pkg/dump/mysql"
	pgdump "github.com/kunitsucom/ddlctl/pkg/dump/postgres"
	apperr "github.com/kunitsucom/ddlctl/pkg/errors"
	"github.com/kunitsucom/ddlctl/pkg/internal/config"
)

func Dump(ctx context.Context, args []string) error {
	if _, err := config.Load(ctx); err != nil {
		return errorz.Errorf("config.Load: %w", err)
	}

	ddl, err := DumpDDL(ctx, config.Dialect(), args[0])
	if err != nil {
		return errorz.Errorf("diff: %w", err)
	}

	if _, err := io.WriteString(os.Stdout, ddl); err != nil {
		return errorz.Errorf("io.WriteString: %w", err)
	}

	return nil
}

//nolint:cyclop
func DumpDDL(ctx context.Context, dialect string, dsn string) (ddl string, err error) {
	switch dialect {
	case _mysql:
		db, err := sqlz.OpenContext(ctx, _mysql, dsn)
		if err != nil {
			return "", errorz.Errorf("sqlz.OpenContext: %w", err)
		}
		defer func() {
			if cerr := db.Close(); err == nil && cerr != nil {
				err = errorz.Errorf("db.Close: %w", cerr)
			}
		}()

		ddl, err := mydump.ShowCreateAllTables(ctx, db)
		if err != nil {
			return "", errorz.Errorf("pgutil.ShowCreateAllTables: %w", err)
		}

		return ddl, nil
	case _postgres:
		db, err := sqlz.OpenContext(ctx, _postgres, dsn)
		if err != nil {
			return "", errorz.Errorf("sqlz.OpenContext: %w", err)
		}
		defer func() {
			if cerr := db.Close(); err == nil && cerr != nil {
				err = errorz.Errorf("db.Close: %w", cerr)
			}
		}()

		ddl, err := pgdump.ShowCreateAllTables(ctx, db)
		if err != nil {
			return "", errorz.Errorf("pgutil.ShowCreateAllTables: %w", err)
		}

		return ddl, nil
	case _cockroachdb:
		db, err := sqlz.OpenContext(ctx, _postgres /* cockroachdb's driver is postgres */, dsn)
		if err != nil {
			return "", errorz.Errorf("sqlz.OpenContext: %w", err)
		}
		defer func() {
			if cerr := db.Close(); err == nil && cerr != nil {
				err = errorz.Errorf("db.Close: %w", cerr)
			}
		}()

		ddl, err := crdbdump.ShowCreateAllTables(ctx, db)
		if err != nil {
			return "", errorz.Errorf("crdbutil.ShowCreateAllTables: %w", err)
		}

		return ddl, nil
	default:
		return "", errorz.Errorf("dialect=%s: %w", dialect, apperr.ErrNotSupported)
	}
}
