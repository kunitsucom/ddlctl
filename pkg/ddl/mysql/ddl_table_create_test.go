package mysql

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
				{Name: &Ident{Name: "name", Raw: "name"}, DataType: &DataType{Name: "VARYING", Expr: &Expr{[]*Ident{{Name: "255", Raw: "255"}}}}},
			},
			Options: []*Option{
				{Name: "ENGINE", Value: &Ident{Name: "InnoDB", Raw: "InnoDB"}},
				{Name: "DEFAULT CHARSET", Value: &Ident{Name: "utf8mb4", Raw: "utf8mb4"}},
				{Name: "COLLATE", Value: &Ident{Name: "utf8mb4_0900_ai_ci", Raw: "utf8mb4_0900_ai_ci"}},
			},
		}
		expected := `-- test comment content
CREATE TABLE test (
    id INTEGER,
    name VARYING(255)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
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
