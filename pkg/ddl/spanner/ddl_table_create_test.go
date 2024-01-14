package spanner

import (
	"testing"

	"github.com/kunitsucom/util.go/testing/assert"
)

func TestCreateTableStmt_String(t *testing.T) {
	t.Parallel()

	t.Run("success,", func(t *testing.T) {
		t.Parallel()

		stmt := &CreateTableStmt{
			Comment: "test comment content",
			Indent:  "  ",
			Name:    &ObjectName{Name: &Ident{Name: "test", Raw: "test"}},
			Columns: []*Column{
				{Name: &Ident{Name: "id", Raw: "id"}, DataType: &DataType{Name: "INTEGER"}},
				{Name: &Ident{Name: "name", Raw: "name"}, DataType: &DataType{Name: "VARYING", Expr: &Expr{Idents: []*Ident{NewRawIdent("255")}}}},
			},
			Options: []*Option{
				{Name: "PRIMARY KEY", Value: &Expr{Idents: []*Ident{NewRawIdent("("), NewRawIdent("id"), NewRawIdent(")")}}},
			},
		}
		expected := `-- test comment content
CREATE TABLE test (
    id INTEGER,
    name VARYING(255)
) PRIMARY KEY (id);
`

		actual := stmt.String()
		assert.Equal(t, expected, actual)

		t.Logf("✅: %s: stmt: %#v", t.Name(), stmt)
	})
}

func TestCreateTableStmt_GetNameForDiff(t *testing.T) {
	t.Parallel()

	t.Run("success,", func(t *testing.T) {
		t.Parallel()

		stmt := &CreateTableStmt{Name: &ObjectName{Name: &Ident{Name: "test", QuotationMark: `"`, Raw: `"test"`}}}
		expected := "test"
		actual := stmt.GetNameForDiff()

		assert.Equal(t, expected, actual)

		t.Logf("✅: %s: stmt: %#v", t.Name(), stmt)
	})
}
