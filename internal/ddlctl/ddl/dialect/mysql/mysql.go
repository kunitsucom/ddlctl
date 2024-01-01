package mysql

import (
	"io"

	errorz "github.com/kunitsucom/util.go/errors"

	ddlast "github.com/kunitsucom/ddlctl/internal/ddlctl/ddl"
	"github.com/kunitsucom/ddlctl/internal/logs"
	"github.com/kunitsucom/ddlctl/pkg/errors"
)

const (
	Dialect       = "mysql"
	CommentPrefix = "--"
	Quotation     = "`"
)

func Fprint(w io.Writer, ddl *ddlast.DDL) error {
	var buf string

	for _, header := range ddl.Header {
		fprintComment(&buf, "", header)
	}

	for _, statement := range ddl.Stmts {
		buf += "\n"
		switch stmt := statement.(type) {
		case *ddlast.CreateTableStmt:
			fprintCreateTable(&buf, ddl.Indent, stmt)
		case *ddlast.CreateIndexStmt:
			fprintCreateIndex(&buf, ddl.Indent, stmt)
		default:
			logs.Warn.Printf("unknown statement type: %T: %v", stmt, errors.ErrNotSupported)
			continue
		}
	}

	if _, err := io.WriteString(w, buf); err != nil {
		return errorz.Errorf("io.WriteString: %w", err)
	}
	return nil
}

func fprintComment(buf *string, indent string, comment string) {
	if comment == "" {
		*buf += indent + CommentPrefix + "\n"
		return
	}

	*buf += indent + CommentPrefix + " " + comment + "\n"
	return //nolint:gosimple
}
