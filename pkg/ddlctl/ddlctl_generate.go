package ddlctl

import (
	"context"
	"io"
	"os"
	"path/filepath"

	errorz "github.com/kunitsucom/util.go/errors"

	apperr "github.com/kunitsucom/ddlctl/pkg/errors"
	"github.com/kunitsucom/ddlctl/pkg/internal/config"
	"github.com/kunitsucom/ddlctl/pkg/internal/generator"
	"github.com/kunitsucom/ddlctl/pkg/internal/generator/dialect/mysql"
	"github.com/kunitsucom/ddlctl/pkg/internal/generator/dialect/postgres"
	"github.com/kunitsucom/ddlctl/pkg/internal/generator/dialect/spanner"
	ddlctlgo "github.com/kunitsucom/ddlctl/pkg/internal/lang/go"
	"github.com/kunitsucom/ddlctl/pkg/internal/logs"
)

func Generate(ctx context.Context, _ []string) error {
	if _, err := config.Load(ctx); err != nil {
		return errorz.Errorf("config.Load: %w", err)
	}

	src := config.Source()
	logs.Info.Printf("source: %s", src)

	language := config.Language()
	logs.Info.Printf("language: %s", language)

	ddl, err := Parse(ctx, language, src)
	if err != nil {
		return errorz.Errorf("parse: %w", err)
	}

	if info, err := os.Stat(config.Destination()); err == nil && info.IsDir() {
		for _, stmt := range ddl.Stmts {
			dst := filepath.Join(config.Destination(), filepath.Base(stmt.GetSourceFile())) + ".gen.sql"
			logs.Info.Printf("destination: %s", dst)

			f, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
			if err != nil {
				return errorz.Errorf("os.OpenFile: %w", err)
			}

			if err := Fprint(
				f,
				config.Dialect(),
				&generator.DDL{
					Header: ddl.Header,
					Indent: ddl.Indent,
					Stmts:  []generator.Stmt{stmt},
				},
			); err != nil {
				return errorz.Errorf("fprint: %w", err)
			}
		}
		return nil
	}

	dst := config.Destination()
	logs.Info.Printf("destination: %s", dst)

	f, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		return errorz.Errorf("os.OpenFile: %w", err)
	}

	if err := Fprint(f, config.Dialect(), ddl); err != nil {
		return errorz.Errorf("fprint: %w", err)
	}
	return nil
}

func Parse(ctx context.Context, language string, src string) (*generator.DDL, error) {
	switch language {
	case ddlctlgo.Language:
		ddl, err := ddlctlgo.Parse(ctx, src)
		if err != nil {
			return nil, errorz.Errorf("ddlgengo.Parse: %w", err)
		}
		return ddl, nil
	default:
		return nil, errorz.Errorf("language=%s: %w", language, apperr.ErrNotSupported)
	}
}

func Fprint(w io.Writer, dialect string, ddl *generator.DDL) error {
	switch dialect {
	case spanner.Dialect:
		if err := spanner.Fprint(w, ddl); err != nil {
			return errorz.Errorf("spanner.Fprint: %w", err)
		}
		return nil
	case postgres.Dialect, "cockroachdb":
		if err := postgres.Fprint(w, ddl); err != nil {
			return errorz.Errorf("postgres.Fprint: %w", err)
		}
		return nil
	case mysql.Dialect:
		if err := mysql.Fprint(w, ddl); err != nil {
			return errorz.Errorf("mysql.Fprint: %w", err)
		}
		return nil
	case "":
		return errorz.Errorf("dialect=%s: %w", dialect, apperr.ErrDialectIsEmpty)
	default:
		return errorz.Errorf("dialect=%s: %w", dialect, apperr.ErrNotSupported)
	}
}
