package postgres

import (
	"fmt"

	filepathz "github.com/kunitsucom/util.go/path/filepath"

	ddlast "github.com/kunitsucom/ddlctl/pkg/internal/generator"
)

//nolint:cyclop,funlen
func fprintCreateIndex(buf *string, _ string, stmt *ddlast.CreateIndexStmt) {
	// source
	if stmt.SourceFile != "" {
		fprintComment(buf, "", fmt.Sprintf("source: %s:%d", filepathz.Short(stmt.SourceFile), stmt.SourceLine))
	}

	// comments
	for _, comment := range stmt.Comments {
		fprintComment(buf, "", comment)
	}

	// CREATE INDEX
	*buf += stmt.CreateIndex

	*buf += ";\n"

	return //nolint:gosimple
}
