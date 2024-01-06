package ddlctl

import (
	"context"
	"io"
	"os"
	"regexp"
	"strings"

	sqlz "github.com/kunitsucom/util.go/database/sql"
	errorz "github.com/kunitsucom/util.go/errors"
	crdbddl "github.com/kunitsucom/util.go/exp/database/sql/ddl/cockroachdb"
	pgddl "github.com/kunitsucom/util.go/exp/database/sql/ddl/postgres"
	crdbutil "github.com/kunitsucom/util.go/exp/database/sql/util/cockroachdb"
	myutil "github.com/kunitsucom/util.go/exp/database/sql/util/mysql"
	pgutil "github.com/kunitsucom/util.go/exp/database/sql/util/postgres"
	osz "github.com/kunitsucom/util.go/os"

	"github.com/kunitsucom/ddlctl/internal/config"
	"github.com/kunitsucom/ddlctl/internal/logs"
	apperr "github.com/kunitsucom/ddlctl/pkg/errors"
)

const (
	_mysql       = "mysql"
	_postgres    = "postgres"
	_cockroachdb = "cockroachdb"
)

func Diff(ctx context.Context, args []string) error {
	if _, err := config.Load(ctx); err != nil {
		return errorz.Errorf("config.Load: %w", err)
	}

	if len(args) != 2 {
		return errorz.Errorf("args=%v: %w", args, apperr.ErrTwoArgumentsRequired)
	}

	left, right, err := resolve(ctx, config.Dialect(), args[0], args[1])
	if err != nil {
		return errorz.Errorf("resolve: %w", err)
	}

	if err := diff(os.Stdout, left, right); err != nil {
		return errorz.Errorf("diff: %w", err)
	}

	return nil
}

var dsnRegex = regexp.MustCompile(`^([a-zA-Z0-9]+):\/\/(.+)$`)

//nolint:cyclop
func resolve(ctx context.Context, dialect, left, right string) (srcDDL string, dstDDL string, err error) {
	var leftIsDSN bool
	switch {
	case dsnRegex.MatchString(left): // NOTE: expect DSN
		ddl, err := dumpCreateStmts(ctx, dialect, left)
		if err != nil {
			return "", "", errorz.Errorf("dumpCreateStmts: %w", err)
		}
		srcDDL = ddl
	case osz.IsFile(left): // NOTE: expect SQL file
		ddlBytes, err := os.ReadFile(left)
		if err != nil {
			return "", "", errorz.Errorf("os.ReadFile: %w", err)
		}
		srcDDL = string(ddlBytes)
	default: // NOTE: expect ddlctl generate format
		ddl, err := generateDDLForDiff(ctx, left)
		if err != nil {
			return "", "", errorz.Errorf("generateDDL: %w", err) // TODO: ddlgen 形式じゃないから無理というエラーに修正する
		}
		srcDDL = ddl
	}

	switch {
	case dsnRegex.MatchString(right): // NOTE: expect DSN
		if leftIsDSN {
			return "", "", errorz.Errorf("left=%s, right=%s: %w", left, right, apperr.ErrBothArgumentsIsDSN) // TODO: define error
		}

		ddl, err := dumpCreateStmts(ctx, dialect, right)
		if err != nil {
			return "", "", errorz.Errorf("dumpCreateStmts: %w", err)
		}
		dstDDL = ddl
	case osz.IsFile(right): // NOTE: expect SQL file
		ddlBytes, err := os.ReadFile(right)
		if err != nil {
			return "", "", errorz.Errorf("os.ReadFile: %w", err)
		}
		dstDDL = string(ddlBytes)
	default: // NOTE: expect ddlctl generate format
		ddl, err := generateDDLForDiff(ctx, right)
		if err != nil {
			return "", "", errorz.Errorf("generateDDL: %w", err) // TODO: ddlgen 形式じゃないから無理というエラーに修正する
		}
		dstDDL = ddl
	}

	return srcDDL, dstDDL, nil
}

//nolint:cyclop
func dumpCreateStmts(ctx context.Context, dialect string, dsn string) (ddl string, err error) {
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

		ddl, err := myutil.ShowCreateAllTables(ctx, db)
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

		ddl, err := pgutil.ShowCreateAllTables(ctx, db)
		if err != nil {
			return "", errorz.Errorf("pgutil.ShowCreateAllTables: %w", err)
		}

		return ddl, nil
	case _cockroachdb:
		db, err := sqlz.OpenContext(ctx, _postgres, dsn)
		if err != nil {
			return "", errorz.Errorf("sqlz.OpenContext: %w", err)
		}
		defer func() {
			if cerr := db.Close(); err == nil && cerr != nil {
				err = errorz.Errorf("db.Close: %w", cerr)
			}
		}()

		ddl, err := crdbutil.ShowCreateAllTables(ctx, db)
		if err != nil {
			return "", errorz.Errorf("crdbutil.ShowCreateAllTables: %w", err)
		}

		return ddl, nil
	default:
		return "", errorz.Errorf("dialect=%s: %w", dialect, apperr.ErrNotSupported)
	}
}

func generateDDLForDiff(ctx context.Context, src string) (string, error) {
	ddl, err := parse(ctx, src)
	if err != nil {
		return "", errorz.Errorf("parse: %w", err)
	}

	b := new(strings.Builder)
	if err := fprint(b, ddl); err != nil {
		return "", errorz.Errorf("fprint: %w", err)
	}

	return b.String(), nil
}

//nolint:cyclop
func diff(out io.Writer, src, dst string) error {
	logs.Debug.Printf("src: %q", src)
	logs.Debug.Printf("dst: %q", dst)

	switch dialect := config.Dialect(); dialect {
	case _mysql:
		// TODO: implement
		return errorz.Errorf("dialect=%s: %w", dialect, apperr.ErrNotSupported)
	case _postgres:
		leftDDL, err := pgddl.NewParser(pgddl.NewLexer(src)).Parse()
		if err != nil {
			return errorz.Errorf("pgddl.NewParser: %w", err)
		}
		rightDDL, err := pgddl.NewParser(pgddl.NewLexer(dst)).Parse()
		if err != nil {
			return errorz.Errorf("pgddl.NewParser: %w", err)
		}

		result, err := pgddl.Diff(leftDDL, rightDDL)
		if err != nil {
			return errorz.Errorf("pgddl.Diff: %w", err)
		}

		if _, err := io.WriteString(out, result.String()); err != nil {
			return errorz.Errorf("io.WriteString: %w", err)
		}

		return nil
	case _cockroachdb:
		leftDDL, err := crdbddl.NewParser(crdbddl.NewLexer(src)).Parse()
		if err != nil {
			return errorz.Errorf("pgddl.NewParser: %w", err)
		}
		rightDDL, err := crdbddl.NewParser(crdbddl.NewLexer(dst)).Parse()
		if err != nil {
			return errorz.Errorf("pgddl.NewParser: %w", err)
		}

		result, err := crdbddl.Diff(leftDDL, rightDDL)
		if err != nil {
			return errorz.Errorf("pgddl.Diff: %w", err)
		}

		if _, err := io.WriteString(out, result.String()); err != nil {
			return errorz.Errorf("io.WriteString: %w", err)
		}

		return nil
	case "":
		return errorz.Errorf("dialect=%s: %w", dialect, apperr.ErrDialectIsEmpty)
	default:
		return errorz.Errorf("dialect=%s: %w", dialect, apperr.ErrNotSupported)
	}
}
