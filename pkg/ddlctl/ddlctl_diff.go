package ddlctl

import (
	"context"
	"os"
	"regexp"
	"strings"

	sqlz "github.com/kunitsucom/util.go/database/sql"
	errorz "github.com/kunitsucom/util.go/errors"
	crdbddl "github.com/kunitsucom/util.go/exp/database/sql/ddl/cockroachdb"
	pgddl "github.com/kunitsucom/util.go/exp/database/sql/ddl/postgres"
	osz "github.com/kunitsucom/util.go/os"

	"github.com/kunitsucom/ddlctl/internal/config"
	pgddlgen "github.com/kunitsucom/ddlctl/internal/ddlctl/ddl/dialect/postgres"
	"github.com/kunitsucom/ddlctl/internal/logs"
	apperr "github.com/kunitsucom/ddlctl/pkg/errors"
)

const (
	_postgres    = "postgres"
	_cockroachdb = "cockroachdb"
)

func Diff(ctx context.Context, args []string) error {
	if _, err := config.Load(ctx); err != nil {
		return errorz.Errorf("config.Load: %w", err)
	}

	left, right, err := resolve(ctx, config.Dialect(), args[0], args[1])
	if err != nil {
		return errorz.Errorf("resolve: %w", err)
	}

	if err := diff(left, right); err != nil {
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
			return "", "", errorz.Errorf("sqlz.OpenContext: %w", err)
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

func dumpCreateStmts(ctx context.Context, dialect string, dsn string) (ddl string, err error) {
	switch dialect {
	case _cockroachdb:
		db, err := sqlz.OpenContext(ctx, _postgres, dsn)
		if err != nil {
			return "", errorz.Errorf("sqlz.OpenContext: %w", err)
		}
		defer db.Close()

		type CreateTableStatement struct {
			CreateStatement string `db:"create_statement"`
		}
		v := new([]*CreateTableStatement)
		if err := sqlz.NewDB(db).QueryContext(ctx, v, "SHOW CREATE ALL TABLES;"); err != nil {
			return "", errorz.Errorf("sqlz.NewDB.QueryContext: %w", err)
		}
		for _, stmt := range *v {
			ddl += stmt.CreateStatement
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
func diff(src, dst string) error {
	logs.Debug.Printf("src: %q", src)
	logs.Debug.Printf("dst: %q", dst)

	switch dialect := config.Dialect(); dialect {
	case pgddlgen.Dialect:
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

		os.Stdout.WriteString(result.String())

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

		os.Stdout.WriteString(result.String())

		return nil
	case "":
		return errorz.Errorf("dialect=%s: %w", dialect, apperr.ErrDialectIsEmpty)
	default:
		return errorz.Errorf("dialect=%s: %w", dialect, apperr.ErrNotSupported)
	}
}
