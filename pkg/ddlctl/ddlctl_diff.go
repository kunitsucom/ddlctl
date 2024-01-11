package ddlctl

import (
	"context"
	"io"
	"os"
	"strings"

	errorz "github.com/kunitsucom/util.go/errors"
	osz "github.com/kunitsucom/util.go/os"

	crdbddl "github.com/kunitsucom/ddlctl/pkg/ddl/cockroachdb"
	myddl "github.com/kunitsucom/ddlctl/pkg/ddl/mysql"
	pgddl "github.com/kunitsucom/ddlctl/pkg/ddl/postgres"

	apperr "github.com/kunitsucom/ddlctl/pkg/errors"
	"github.com/kunitsucom/ddlctl/pkg/internal/config"
	"github.com/kunitsucom/ddlctl/pkg/internal/logs"
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

	if err := DiffDDL(os.Stdout, config.Dialect(), left, right); err != nil {
		return errorz.Errorf("diff: %w", err)
	}

	return nil
}

//nolint:cyclop
func resolve(ctx context.Context, dialect, left, right string) (srcDDL string, dstDDL string, err error) {
	switch {
	case osz.IsFile(left): // NOTE: expect SQL file
		ddlBytes, err := os.ReadFile(left)
		if err != nil {
			return "", "", errorz.Errorf("os.ReadFile: %w", err)
		}
		srcDDL = string(ddlBytes)
	case osz.Exists(left): // NOTE: expect ddlctl generate format
		ddl, err := generateDDLForDiff(ctx, left)
		if err != nil {
			return "", "", errorz.Errorf("generateDDL: %w", err) // TODO: ddlgen 形式じゃないから無理というエラーに修正する
		}
		srcDDL = ddl
	default: // NOTE: expect DSN
		ddl, err := DumpDDL(ctx, dialect, left)
		if err != nil {
			return "", "", errorz.Errorf("dumpCreateStmts: %w", err)
		}
		srcDDL = ddl
	}

	switch {
	case osz.IsFile(right): // NOTE: expect SQL file
		ddlBytes, err := os.ReadFile(right)
		if err != nil {
			return "", "", errorz.Errorf("os.ReadFile: %w", err)
		}
		dstDDL = string(ddlBytes)
	case osz.Exists(right): // NOTE: expect ddlctl generate format
		ddl, err := generateDDLForDiff(ctx, right)
		if err != nil {
			return "", "", errorz.Errorf("generateDDL: %w", err) // TODO: ddlgen 形式じゃないから無理というエラーに修正する
		}
		dstDDL = ddl
	default: // NOTE: expect ddlctl generate format
		ddl, err := DumpDDL(ctx, dialect, right)
		if err != nil {
			return "", "", errorz.Errorf("dumpCreateStmts: %w", err)
		}
		dstDDL = ddl
	}

	return srcDDL, dstDDL, nil
}

func generateDDLForDiff(ctx context.Context, src string) (string, error) {
	ddl, err := Parse(ctx, config.Language(), src)
	if err != nil {
		return "", errorz.Errorf("parse: %w", err)
	}

	b := new(strings.Builder)
	if err := Fprint(b, config.Dialect(), ddl); err != nil {
		return "", errorz.Errorf("fprint: %w", err)
	}

	return b.String(), nil
}

//nolint:cyclop,funlen
func DiffDDL(out io.Writer, dialect string, srcDDL string, dstDDL string) error {
	logs.Trace.Printf("src: %q", srcDDL)
	logs.Trace.Printf("dst: %q", dstDDL)

	switch dialect {
	case _mysql:
		leftDDL, err := myddl.NewParser(myddl.NewLexer(srcDDL)).Parse()
		if err != nil {
			return errorz.Errorf("myddl.NewParser: %w", err)
		}
		rightDDL, err := myddl.NewParser(myddl.NewLexer(dstDDL)).Parse()
		if err != nil {
			return errorz.Errorf("myddl.NewParser: %w", err)
		}

		result, err := myddl.Diff(leftDDL, rightDDL)
		if err != nil {
			return errorz.Errorf("myddl.Diff: %w", err)
		}

		if _, err := io.WriteString(out, result.String()); err != nil {
			return errorz.Errorf("io.WriteString: %w", err)
		}

		return nil
	case _postgres:
		leftDDL, err := pgddl.NewParser(pgddl.NewLexer(srcDDL)).Parse()
		if err != nil {
			return errorz.Errorf("pgddl.NewParser: %w", err)
		}
		rightDDL, err := pgddl.NewParser(pgddl.NewLexer(dstDDL)).Parse()
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
		leftDDL, err := crdbddl.NewParser(crdbddl.NewLexer(srcDDL)).Parse()
		if err != nil {
			return errorz.Errorf("pgddl.NewParser: %w", err)
		}
		rightDDL, err := crdbddl.NewParser(crdbddl.NewLexer(dstDDL)).Parse()
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
