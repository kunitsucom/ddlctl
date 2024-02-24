//nolint:testpackage
package spanner

import (
	"bytes"
	"context"
	"io"
	"testing"

	testingz "github.com/kunitsucom/util.go/testing"
	"github.com/kunitsucom/util.go/testing/assert"
	"github.com/kunitsucom/util.go/testing/require"

	ddlast "github.com/kunitsucom/ddlctl/pkg/internal/generator"
	"github.com/kunitsucom/ddlctl/pkg/logs"
)

//nolint:paralleltest
func TestFprint(t *testing.T) {
	t.Run("success,None", func(t *testing.T) {
		ddl := ddlast.NewDDL(context.Background())
		ddl.Stmts = []ddlast.Stmt{
			&ddlast.CreateTableStmt{
				Comments:    []string{"Spans is Spanner test table."},
				CreateTable: "CREATE TABLE Spans",
				Columns: []*ddlast.CreateTableColumn{
					{
						ColumnName:     "Id",
						TypeConstraint: "STRING(64) NOT NULL",
						Comments:       []string{"Id is Spans's Id."},
					},
					{
						ColumnName:     "Name",
						TypeConstraint: "STRING(100) NOT NULL",
						Comments:       []string{"Name is Spans's Name."},
					},
					{
						ColumnName:     "Number",
						TypeConstraint: "INT64 NOT NULL",
					},
					{
						ColumnName:     "Description",
						TypeConstraint: "STRING(1024) NOT NULL",
						Comments:       []string{"Description is Spans's Description."},
					},
					{
						ColumnName:     "CreatedAt",
						TypeConstraint: "TIMESTAMP NOT NULL OPTIONS (allow_commit_timestamp=true)",
						Comments:       []string{"CreatedAt is Spans's CreatedAt."},
					},
					{
						ColumnName:     "UpdatedAt",
						TypeConstraint: "TIMESTAMP NOT NULL OPTIONS (allow_commit_timestamp=true)",
						Comments:       []string{"UpdatedAt is Spans's UpdatedAt."},
					},
				},
				Constraints: []*ddlast.CreateTableConstraint{
					{
						Constraint: "CONSTRAINT NumberGteZero CHECK(Number >= 0)",
					},
					{
						Comments:   []string{"CREATE TABLE CONSTRAINT COMMENT"},
						Constraint: "CONSTRAINT CreateBeforeUpdate CHECK(CreatedAt <= UpdatedAt)",
					},
				},
				Options: []*ddlast.CreateTableOption{
					{
						Option: "PRIMARY KEY (Id)",
					},
					{
						Comments: []string{"CREATE TABLE OPTION COMMENT: If SpanParents record is deleted, Spans record is deleted."},
						Option:   "INTERLEAVE IN PARENT SpanParents ON DELETE CASCADE",
					},
				},
			},
		}

		const expected = `-- Code generated by ddlctl. DO NOT EDIT.
--

-- Spans is Spanner test table.
CREATE TABLE Spans (
    -- Id is Spans's Id.
    ` + "`Id`" + `          STRING(64) NOT NULL,
    -- Name is Spans's Name.
    ` + "`Name`" + `        STRING(100) NOT NULL,
    ` + "`Number`" + `      INT64 NOT NULL,
    -- Description is Spans's Description.
    ` + "`Description`" + ` STRING(1024) NOT NULL,
    -- CreatedAt is Spans's CreatedAt.
    ` + "`CreatedAt`" + `   TIMESTAMP NOT NULL OPTIONS (allow_commit_timestamp=true),
    -- UpdatedAt is Spans's UpdatedAt.
    ` + "`UpdatedAt`" + `   TIMESTAMP NOT NULL OPTIONS (allow_commit_timestamp=true),
    CONSTRAINT NumberGteZero CHECK(Number >= 0),
    -- CREATE TABLE CONSTRAINT COMMENT
    CONSTRAINT CreateBeforeUpdate CHECK(CreatedAt <= UpdatedAt)
)
PRIMARY KEY (Id),
-- CREATE TABLE OPTION COMMENT: If SpanParents record is deleted, Spans record is deleted.
INTERLEAVE IN PARENT SpanParents ON DELETE CASCADE;
`

		buf := bytes.NewBuffer(nil)
		if err := Fprint(buf, ddl); err != nil {
			t.Fatalf("failed to Fprint: %+v", err)
		}
		actual := buf.String()

		assert.Equal(t, expected, actual)
	})

	t.Run("failure,Write", func(t *testing.T) {
		ddl := ddlast.NewDDL(context.Background())
		ddl.Stmts = []ddlast.Stmt{
			nil,
		}

		w := &testingz.Writer{WriteFunc: func(p []byte) (int, error) {
			return 0, io.ErrUnexpectedEOF
		}}

		backup := logs.Warn
		t.Cleanup(func() { logs.Warn = backup })
		logs.Warn = logs.NewDebug()

		err := Fprint(w, ddl)
		require.Error(t, err)
		require.ErrorIs(t, err, io.ErrUnexpectedEOF)
	})
}
