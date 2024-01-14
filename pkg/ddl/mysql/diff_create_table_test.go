package mysql

import (
	"testing"

	"github.com/kunitsucom/util.go/testing/assert"
	"github.com/kunitsucom/util.go/testing/require"

	"github.com/kunitsucom/ddlctl/pkg/ddl"
)

//nolint:paralleltest,tparallel
func TestDiffCreateTable(t *testing.T) {
	t.Run("failure,ddl.ErrNoDifference", func(t *testing.T) {
		t.Parallel()

		before := `CREATE TABLE "users" (id VARCHAR(36) NOT NULL, group_id VARCHAR(36) NOT NULL REFERENCES "groups" ("id"), "name" VARCHAR(255) NOT NULL UNIQUE, description TEXT, PRIMARY KEY ("id"));`
		beforeDDL, err := NewParser(NewLexer(before)).Parse()
		require.NoError(t, err)

		after := `CREATE TABLE "users" (id VARCHAR(36) NOT NULL, group_id VARCHAR(36) NOT NULL REFERENCES "groups" ("id"), "name" VARCHAR(255) NOT NULL UNIQUE, description TEXT, PRIMARY KEY ("id"));`

		afterDDL, err := NewParser(NewLexer(after)).Parse()
		require.NoError(t, err)

		actual, err := DiffCreateTable(
			beforeDDL.Stmts[0].(*CreateTableStmt),
			afterDDL.Stmts[0].(*CreateTableStmt),
			DiffCreateTableUseAlterTableAddConstraintNotValid(false),
		)

		assert.ErrorIs(t, err, ddl.ErrNoDifference)
		assert.Nil(t, actual)

		t.Logf("✅: %s: actual: %%#v: \n%#v", t.Name(), actual)
		t.Logf("✅: %s: actual: %%s: \n%s", t.Name(), actual)
	})

	t.Run("success,ADD_COLUMN", func(t *testing.T) {
		t.Parallel()

		before := `CREATE TABLE "users" (id VARCHAR(36) NOT NULL, group_id VARCHAR(36) NOT NULL REFERENCES "groups" ("id"), "name" VARCHAR(255) NOT NULL UNIQUE, description TEXT, PRIMARY KEY ("id"));`
		beforeDDL, err := NewParser(NewLexer(before)).Parse()
		require.NoError(t, err)

		after := `CREATE TABLE "users" (id VARCHAR(36) NOT NULL, group_id VARCHAR(36) NOT NULL REFERENCES "groups" ("id"), "name" VARCHAR(255) NOT NULL UNIQUE, "age" INTEGER DEFAULT 0 NOT NULL CHECK ("age" >= 0), description TEXT, PRIMARY KEY ("id"));`

		afterDDL, err := NewParser(NewLexer(after)).Parse()
		require.NoError(t, err)

		actual, err := DiffCreateTable(
			beforeDDL.Stmts[0].(*CreateTableStmt),
			afterDDL.Stmts[0].(*CreateTableStmt),
			DiffCreateTableUseAlterTableAddConstraintNotValid(false),
		)

		expectedStr := `-- -
-- +"age" INTEGER NOT NULL DEFAULT 0
ALTER TABLE "users" ADD COLUMN "age" INTEGER NOT NULL DEFAULT 0;
-- -
-- +CONSTRAINT users_age_check CHECK ("age" >= 0)
ALTER TABLE "users" ADD CONSTRAINT users_age_check CHECK ("age" >= 0);
`

		assert.NoError(t, err)
		assert.Equal(t, expectedStr, actual.String())

		t.Logf("✅: %s: actual: %%#v:\n%#v", t.Name(), actual)
	})

	t.Run("success,DROP_COLUMN", func(t *testing.T) {
		t.Parallel()

		before := `CREATE TABLE "users" (id VARCHAR(36) NOT NULL, group_id VARCHAR(36) NOT NULL REFERENCES "groups" ("id"), "name" VARCHAR(255) NOT NULL UNIQUE, "age" INTEGER DEFAULT 0 NOT NULL CHECK ("age" >= 0), description TEXT, PRIMARY KEY ("id"));`
		beforeDDL, err := NewParser(NewLexer(before)).Parse()
		require.NoError(t, err)

		after := `CREATE TABLE "users" (id VARCHAR(36) NOT NULL, group_id VARCHAR(36) NOT NULL REFERENCES "groups" ("id"), "name" VARCHAR(255) NOT NULL, description TEXT, PRIMARY KEY ("id"));`

		afterDDL, err := NewParser(NewLexer(after)).Parse()
		require.NoError(t, err)

		actual, err := DiffCreateTable(
			beforeDDL.Stmts[0].(*CreateTableStmt),
			afterDDL.Stmts[0].(*CreateTableStmt),
			DiffCreateTableUseAlterTableAddConstraintNotValid(false),
		)

		expectedStr := `-- -UNIQUE KEY users_unique_name (name)
-- +
DROP INDEX users_unique_name;
-- -CONSTRAINT users_age_check CHECK ("age" >= 0)
-- +
ALTER TABLE "users" DROP CONSTRAINT users_age_check;
-- -"age" INTEGER NOT NULL DEFAULT 0
-- +
ALTER TABLE "users" DROP COLUMN "age";
`

		assert.NoError(t, err)
		assert.Equal(t, expectedStr, actual.String())

		t.Logf("✅: %s: actual: %%#v:\n%#v", t.Name(), actual)
	})

	t.Run("success,ALTER_COLUMN_SET_DATA_TYPE", func(t *testing.T) {
		t.Parallel()

		before := `CREATE TABLE "users" (id VARCHAR(36) NOT NULL, group_id VARCHAR(36) NOT NULL REFERENCES "groups" ("id"), "name" VARCHAR(255) NOT NULL, "age" INT DEFAULT 0 CHECK ("age" >= 0), description TEXT, PRIMARY KEY ("id"));`
		beforeDDL, err := NewParser(NewLexer(before)).Parse()
		require.NoError(t, err)

		after := `CREATE TABLE "users" (id VARCHAR(36) NOT NULL, group_id VARCHAR(36) NOT NULL REFERENCES "groups" ("id"), "name" TEXT NOT NULL UNIQUE, "age" BIGINT DEFAULT 0 CHECK ("age" >= 0), description TEXT, PRIMARY KEY ("id"));`

		afterDDL, err := NewParser(NewLexer(after)).Parse()
		require.NoError(t, err)

		actual, err := DiffCreateTable(
			beforeDDL.Stmts[0].(*CreateTableStmt),
			afterDDL.Stmts[0].(*CreateTableStmt),
			DiffCreateTableUseAlterTableAddConstraintNotValid(false),
		)

		expectedStr := `-- -"name" VARCHAR(255) NOT NULL
-- +"name" TEXT NOT NULL
ALTER TABLE "users" MODIFY "name" TEXT NOT NULL;
-- -"age" INT DEFAULT 0
-- +"age" BIGINT DEFAULT 0
ALTER TABLE "users" MODIFY "age" BIGINT;
-- -
-- +UNIQUE KEY users_unique_name (name)
CREATE UNIQUE INDEX users_unique_name ON "users" ("name");
`

		assert.NoError(t, err)
		assert.Equal(t, expectedStr, actual.String())

		t.Logf("✅: %s: actual: %%#v:\n%#v", t.Name(), actual)
	})

	t.Run("success,ALTER_COLUMN_DROP_DEFAULT", func(t *testing.T) {
		before := `CREATE TABLE "users" (id VARCHAR(36) NOT NULL, group_id VARCHAR(36) NOT NULL REFERENCES "groups" ("id"), "name" VARCHAR(255) NOT NULL UNIQUE, "age" INT DEFAULT 0 CHECK ("age" >= 0), description TEXT, PRIMARY KEY ("id"));`
		beforeDDL, err := NewParser(NewLexer(before)).Parse()
		require.NoError(t, err)

		after := `CREATE TABLE "users" (id VARCHAR(36) NOT NULL, group_id VARCHAR(36) NOT NULL REFERENCES "groups" ("id"), "name" VARCHAR(255) NOT NULL UNIQUE, "age" INT CHECK ("age" >= 0), description TEXT, PRIMARY KEY ("id"));`
		afterDDL, err := NewParser(NewLexer(after)).Parse()
		require.NoError(t, err)

		expectedStr := `-- -"age" INT DEFAULT 0
-- +"age" INT
ALTER TABLE "users" ALTER "age" DROP DEFAULT;
`

		actual, err := DiffCreateTable(
			beforeDDL.Stmts[0].(*CreateTableStmt),
			afterDDL.Stmts[0].(*CreateTableStmt),
			DiffCreateTableUseAlterTableAddConstraintNotValid(false),
		)
		assert.NoError(t, err)
		assert.Equal(t, expectedStr, actual.String())

		t.Logf("✅: %s: actual: %%#v:\n%#v", t.Name(), actual)
	})

	t.Run("success,ALTER_COLUMN_SET_DEFAULT", func(t *testing.T) {
		before := `CREATE TABLE "users" (id VARCHAR(36) NOT NULL, group_id VARCHAR(36) NOT NULL REFERENCES "groups" ("id"), "name" VARCHAR(255) NOT NULL UNIQUE, "age" INT CHECK ("age" >= 0), description TEXT, PRIMARY KEY ("id"));`
		beforeDDL, err := NewParser(NewLexer(before)).Parse()
		require.NoError(t, err)

		after := `CREATE TABLE "users" (id VARCHAR(36) NOT NULL, group_id VARCHAR(36) NOT NULL REFERENCES "groups" ("id"), "name" VARCHAR(255) NOT NULL UNIQUE, "age" INT DEFAULT 0 CHECK ("age" <> 0), description TEXT, PRIMARY KEY (id));`
		afterDDL, err := NewParser(NewLexer(after)).Parse()
		require.NoError(t, err)

		expectedStr := `-- -"age" INT
-- +"age" INT DEFAULT 0
ALTER TABLE "users" ALTER "age" SET DEFAULT 0;
-- -CONSTRAINT users_age_check CHECK ("age" >= 0)
-- +
ALTER TABLE "users" DROP CONSTRAINT users_age_check;
-- -
-- +CONSTRAINT users_age_check CHECK ("age" <> 0)
ALTER TABLE "users" ADD CONSTRAINT users_age_check CHECK ("age" <> 0);
`

		actual, err := DiffCreateTable(
			beforeDDL.Stmts[0].(*CreateTableStmt),
			afterDDL.Stmts[0].(*CreateTableStmt),
			DiffCreateTableUseAlterTableAddConstraintNotValid(false),
		)
		assert.NoError(t, err)
		assert.Equal(t, expectedStr, actual.String())

		t.Logf("✅: %s: actual: %%#v:\n%#v", t.Name(), actual)
	})

	t.Run("success,ALTER_TABLE_RENAME_TO", func(t *testing.T) {
		t.Parallel()

		before := `CREATE TABLE "public.users" (id VARCHAR(36) NOT NULL, group_id VARCHAR(36) NOT NULL REFERENCES "groups" ("id"), "name" VARCHAR(255) NOT NULL UNIQUE, "age" INT DEFAULT 0 CHECK ("age" >= 0), description TEXT, PRIMARY KEY ("id"));`
		beforeDDL, err := NewParser(NewLexer(before)).Parse()
		require.NoError(t, err)

		after := `CREATE TABLE "app_users" (id VARCHAR(36) NOT NULL, group_id VARCHAR(36) NOT NULL REFERENCES "groups" ("id"), "name" VARCHAR(255) NOT NULL UNIQUE, "age" INT DEFAULT 0 CHECK ("age" >= 0), description TEXT, PRIMARY KEY ("id"));`
		afterDDL, err := NewParser(NewLexer(after)).Parse()
		require.NoError(t, err)

		expectedStr := `-- -public.users
-- +public.app_users
ALTER TABLE "public.users" RENAME TO "public.app_users";
-- -CONSTRAINT users_group_id_fkey FOREIGN KEY (group_id) REFERENCES "groups" ("id")
-- +
ALTER TABLE "public.app_users" DROP CONSTRAINT users_group_id_fkey;
-- -UNIQUE KEY users_unique_name (name)
-- +
DROP INDEX public.users_unique_name;
-- -CONSTRAINT users_age_check CHECK ("age" >= 0)
-- +
ALTER TABLE "public.app_users" DROP CONSTRAINT users_age_check;
-- -
-- +CONSTRAINT app_users_group_id_fkey FOREIGN KEY (group_id) REFERENCES "groups" ("id")
ALTER TABLE "public.app_users" ADD CONSTRAINT app_users_group_id_fkey FOREIGN KEY (group_id) REFERENCES "groups" ("id");
-- -
-- +UNIQUE KEY app_users_unique_name (name)
CREATE UNIQUE INDEX public.app_users_unique_name ON "public.app_users" ("name");
-- -
-- +CONSTRAINT app_users_age_check CHECK ("age" >= 0)
ALTER TABLE "public.app_users" ADD CONSTRAINT app_users_age_check CHECK ("age" >= 0);
`

		actual, err := DiffCreateTable(
			beforeDDL.Stmts[0].(*CreateTableStmt),
			afterDDL.Stmts[0].(*CreateTableStmt),
			DiffCreateTableUseAlterTableAddConstraintNotValid(false),
		)
		assert.NoError(t, err)
		assert.Equal(t, expectedStr, actual.String())

		t.Logf("✅: %s: actual: %%#v: \n%#v", t.Name(), actual)
		t.Logf("✅: %s: actual: %%s: \n%s", t.Name(), actual)
	})

	t.Run("success,SET_NOT_NULL", func(t *testing.T) {
		t.Parallel()

		before := `CREATE TABLE "users" (id VARCHAR(36) NOT NULL, group_id VARCHAR(36) NOT NULL REFERENCES "groups" ("id"), "name" VARCHAR(255) NOT NULL UNIQUE, "age" INT DEFAULT 0 CHECK ("age" >= 0), description TEXT, PRIMARY KEY ("id"));`
		beforeDDL, err := NewParser(NewLexer(before)).Parse()
		require.NoError(t, err)

		after := `CREATE TABLE "users" (id VARCHAR(36) NOT NULL, group_id VARCHAR(36) NOT NULL REFERENCES "groups" ("id"), "name" VARCHAR(255) NOT NULL UNIQUE, "age" INTEGER DEFAULT 0 NOT NULL CHECK ("age" >= 0), description TEXT, PRIMARY KEY ("id"));`
		afterDDL, err := NewParser(NewLexer(after)).Parse()
		require.NoError(t, err)

		expectedStr := `-- -"age" INT DEFAULT 0
-- +"age" INTEGER NOT NULL DEFAULT 0
ALTER TABLE "users" MODIFY "age" INTEGER NOT NULL;
`

		actual, err := DiffCreateTable(
			beforeDDL.Stmts[0].(*CreateTableStmt),
			afterDDL.Stmts[0].(*CreateTableStmt),
			DiffCreateTableUseAlterTableAddConstraintNotValid(false),
		)
		assert.NoError(t, err)
		assert.Equal(t, expectedStr, actual.String())

		t.Logf("✅: %s: actual: %%#v:\n%#v", t.Name(), actual)
	})

	t.Run("success,DROP_NOT_NULL", func(t *testing.T) {
		t.Parallel()

		before := `CREATE TABLE "users" (id VARCHAR(36) NOT NULL, group_id VARCHAR(36) NOT NULL REFERENCES "groups" ("id"), "name" VARCHAR(255) NOT NULL UNIQUE, "age" INT DEFAULT 0 NOT NULL CHECK ("age" >= 0), description TEXT, PRIMARY KEY ("id"));`
		beforeDDL, err := NewParser(NewLexer(before)).Parse()
		require.NoError(t, err)

		after := `CREATE TABLE "users" (id VARCHAR(36) NOT NULL, group_id VARCHAR(36) NOT NULL REFERENCES "groups" ("id"), "name" VARCHAR(255) NOT NULL UNIQUE, "age" INT DEFAULT 0 CHECK ("age" >= 0), description TEXT, PRIMARY KEY ("id"));`
		afterDDL, err := NewParser(NewLexer(after)).Parse()
		require.NoError(t, err)

		expectedStr := `-- -"age" INT NOT NULL DEFAULT 0
-- +"age" INT DEFAULT 0
ALTER TABLE "users" MODIFY "age" INT;
`

		actual, err := DiffCreateTable(
			beforeDDL.Stmts[0].(*CreateTableStmt),
			afterDDL.Stmts[0].(*CreateTableStmt),
			DiffCreateTableUseAlterTableAddConstraintNotValid(false),
		)
		assert.NoError(t, err)
		assert.Equal(t, expectedStr, actual.String())

		t.Logf("✅: %s: actual: %%#v:\n%#v", t.Name(), actual)
	})

	t.Run("success,DROP_ADD_PRIMARY_KEY", func(t *testing.T) {
		t.Parallel()

		before := `CREATE TABLE "users" (id VARCHAR(36) NOT NULL, group_id VARCHAR(36) NOT NULL REFERENCES "groups" ("id"), "name" VARCHAR(255) NOT NULL UNIQUE, "age" INT DEFAULT 0 NOT NULL CHECK ("age" >= 0), description TEXT, PRIMARY KEY ("id"));`
		beforeDDL, err := NewParser(NewLexer(before)).Parse()
		require.NoError(t, err)

		after := `CREATE TABLE "users" (id VARCHAR(36) NOT NULL, group_id VARCHAR(36) NOT NULL REFERENCES "groups" ("id"), "name" VARCHAR(255) NOT NULL UNIQUE, "age" INT DEFAULT 0 NOT NULL CHECK ("age" >= 0), description TEXT, PRIMARY KEY ("id", name));`
		afterDDL, err := NewParser(NewLexer(after)).Parse()
		require.NoError(t, err)

		expectedStr := `-- -PRIMARY KEY ("id")
-- +
ALTER TABLE "users" DROP PRIMARY KEY;
-- -
-- +PRIMARY KEY ("id", name)
ALTER TABLE "users" ADD PRIMARY KEY ("id", name);
`

		actual, err := DiffCreateTable(
			beforeDDL.Stmts[0].(*CreateTableStmt),
			afterDDL.Stmts[0].(*CreateTableStmt),
			DiffCreateTableUseAlterTableAddConstraintNotValid(false),
		)
		assert.NoError(t, err)
		assert.Equal(t, expectedStr, actual.String())

		t.Logf("✅: %s: actual: %%#v:\n%#v", t.Name(), actual)
	})

	t.Run("success,DROP_ADD_FOREIGN_KEY", func(t *testing.T) {
		t.Parallel()

		before := `CREATE TABLE "users" (id VARCHAR(36) NOT NULL, group_id VARCHAR(36) NOT NULL, "name" VARCHAR(255) NOT NULL UNIQUE, "age" INT DEFAULT 0 NOT NULL CHECK ("age" >= 0), description TEXT, PRIMARY KEY ("id"), CONSTRAINT users_group_id_fkey FOREIGN KEY (group_id) REFERENCES "groups" ("id"));`
		beforeDDL, err := NewParser(NewLexer(before)).Parse()
		require.NoError(t, err)

		after := `CREATE TABLE "users" (id VARCHAR(36) NOT NULL, group_id VARCHAR(36) NOT NULL, "name" VARCHAR(255) NOT NULL UNIQUE, "age" INT DEFAULT 0 NOT NULL CHECK ("age" >= 0), description TEXT, PRIMARY KEY ("id"), CONSTRAINT users_group_id_fkey FOREIGN KEY (group_id, name) REFERENCES "groups" ("id", name));`
		afterDDL, err := NewParser(NewLexer(after)).Parse()
		require.NoError(t, err)

		expectedStr := `-- -CONSTRAINT users_group_id_fkey FOREIGN KEY (group_id) REFERENCES "groups" ("id")
-- +
ALTER TABLE "users" DROP CONSTRAINT users_group_id_fkey;
-- -
-- +CONSTRAINT users_group_id_fkey FOREIGN KEY (group_id, name) REFERENCES "groups" ("id", name)
ALTER TABLE "users" ADD CONSTRAINT users_group_id_fkey FOREIGN KEY (group_id, name) REFERENCES "groups" ("id", name);
`

		actual, err := DiffCreateTable(
			beforeDDL.Stmts[0].(*CreateTableStmt),
			afterDDL.Stmts[0].(*CreateTableStmt),
			DiffCreateTableUseAlterTableAddConstraintNotValid(false),
		)
		assert.NoError(t, err)
		assert.Equal(t, expectedStr, actual.String())

		t.Logf("✅: %s: actual: %%#v:\n%#v", t.Name(), actual)
	})

	t.Run("success,DROP_ADD_UNIQUE", func(t *testing.T) {
		t.Parallel()

		before := `CREATE TABLE "users" (id VARCHAR(36) NOT NULL, group_id VARCHAR(36) NOT NULL, "name" VARCHAR(255) NOT NULL UNIQUE, "age" INT DEFAULT 0 NOT NULL CHECK ("age" >= 0), description TEXT, PRIMARY KEY ("id"));`
		beforeDDL, err := NewParser(NewLexer(before)).Parse()
		require.NoError(t, err)

		after := `CREATE TABLE "users" (id VARCHAR(36) NOT NULL, group_id VARCHAR(36) NOT NULL, "name" VARCHAR(255) NOT NULL UNIQUE, "age" INT DEFAULT 0 NOT NULL CHECK ("age" >= 0), description TEXT, PRIMARY KEY ("id"), UNIQUE KEY users_unique_name ("id", name));`
		afterDDL, err := NewParser(NewLexer(after)).Parse()
		require.NoError(t, err)

		expectedStr := `DROP INDEX users_unique_name;
CREATE UNIQUE INDEX users_unique_name ON "users" ("id", name);
`

		actual, err := DiffCreateTable(
			beforeDDL.Stmts[0].(*CreateTableStmt),
			afterDDL.Stmts[0].(*CreateTableStmt),
			DiffCreateTableUseAlterTableAddConstraintNotValid(false),
		)
		assert.NoError(t, err)
		assert.Equal(t, expectedStr, actual.String())

		t.Logf("✅: %s: actual: %%#v:\n%#v", t.Name(), actual)
	})

	t.Run("success,ALTER_COLUMN_SET_DEFAULT_OVERWRITE", func(t *testing.T) {
		t.Parallel()

		before := `CREATE TABLE "users" (id VARCHAR(36) NOT NULL, group_id VARCHAR(36) NOT NULL, "name" VARCHAR(255) NOT NULL UNIQUE, "age" INT DEFAULT 0 NOT NULL CHECK ("age" >= 0), description TEXT, PRIMARY KEY ("id"));`
		beforeDDL, err := NewParser(NewLexer(before)).Parse()
		require.NoError(t, err)

		after := `CREATE TABLE "users" (id VARCHAR(36) NOT NULL, group_id VARCHAR(36) NOT NULL, "name" VARCHAR(255) NOT NULL UNIQUE, "age" INT DEFAULT ( (0 + 3) - 1 * 4 / 2 ) NOT NULL CHECK ("age" >= 0), description TEXT, PRIMARY KEY ("id"));`
		afterDDL, err := NewParser(NewLexer(after)).Parse()
		require.NoError(t, err)

		expectedStr := `-- -"age" INT NOT NULL DEFAULT 0
-- +"age" INT NOT NULL DEFAULT ((0 + 3) - 1 * 4 / 2)
ALTER TABLE "users" ALTER "age" SET DEFAULT ((0 + 3) - 1 * 4 / 2);
`

		actual, err := DiffCreateTable(
			beforeDDL.Stmts[0].(*CreateTableStmt),
			afterDDL.Stmts[0].(*CreateTableStmt),
			DiffCreateTableUseAlterTableAddConstraintNotValid(false),
		)
		assert.NoError(t, err)
		assert.Equal(t, expectedStr, actual.String())

		t.Logf("✅: %s: actual: %%#v:\n%#v", t.Name(), actual)
	})

	// 	t.Run("success,ALTER_COLUMN_SET_DEFAULT_complex", func(t *testing.T) {
	// 		t.Parallel()

	// 		before := `CREATE TABLE complex_defaults (
	//     id SERIAL PRIMARY KEY,
	//     created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	//     updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	//     unique_code TEXT,
	//     status TEXT DEFAULT 'pending',
	//     random_number INTEGER DEFAULT FLOOR(RANDOM() * 100)::INTEGER,
	//     json_data JSONB DEFAULT '{}',
	//     calculated_value INTEGER DEFAULT (SELECT COUNT(*) FROM another_table)
	// );
	// `
	// 		beforeDDL, err := NewParser(NewLexer(before)).Parse()
	// 		require.NoError(t, err)

	// 		after := `CREATE TABLE complex_defaults (
	//     id SERIAL PRIMARY KEY,
	//     created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	//     updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	//     unique_code TEXT DEFAULT 'CODE-' || TO_CHAR(NOW(), 'YYYYMMDDHH24MISS') || '-' || LPAD(TO_CHAR(NEXTVAL('seq_complex_default')), 5, '0'),
	//     status TEXT DEFAULT 'pending',
	//     random_number INTEGER DEFAULT FLOOR(RANDOM() * 100)::INTEGER,
	//     json_data JSONB DEFAULT '{}',
	//     calculated_value INTEGER DEFAULT (SELECT COUNT(*) FROM another_table)
	// );
	// `
	// 		afterDDL, err := NewParser(NewLexer(after)).Parse()
	// 		require.NoError(t, err)

	// 		expectedStr := `-- -unique_code TEXT
	// -- +unique_code TEXT DEFAULT 'CODE-' || TO_CHAR(NOW(), 'YYYYMMDDHH24MISS') || '-' || LPAD(TO_CHAR(NEXTVAL('seq_complex_default')), 5, '0')
	// ALTER TABLE complex_defaults ALTER COLUMN unique_code SET DEFAULT 'CODE-' || TO_CHAR(NOW(), 'YYYYMMDDHH24MISS') || '-' || LPAD(TO_CHAR(NEXTVAL('seq_complex_default')), 5, '0');
	// `

	// 		actual, err := DiffCreateTable(
	// 			beforeDDL.Stmts[0].(*CreateTableStmt),
	// 			afterDDL.Stmts[0].(*CreateTableStmt),
	// 			DiffCreateTableUseAlterTableAddConstraintNotValid(false),
	// 		)
	// 		assert.NoError(t, err)
	// 		assert.Equal(t, expectedStr, actual.String())

	// 		t.Logf("✅: %s: actual: %%#v:\n%#v", t.Name(), actual)
	// 	})

	t.Run("success,DiffCreateTableUseAlterTableAddConstraintNotValid", func(t *testing.T) {
		t.Parallel()

		beforeDDL, err := NewParser(NewLexer(`CREATE TABLE "users" (id VARCHAR(36) NOT NULL, group_id VARCHAR(36) NOT NULL REFERENCES "groups" ("id"), "name" VARCHAR(255) NOT NULL UNIQUE, "age" INT DEFAULT 0, description TEXT, PRIMARY KEY ("id"));`)).Parse()
		require.NoError(t, err)

		afterDDL, err := NewParser(NewLexer(`CREATE TABLE "users" (id VARCHAR(36) NOT NULL, group_id VARCHAR(36) NOT NULL REFERENCES "groups" ("id"), "name" VARCHAR(255) NOT NULL UNIQUE, "age" INT DEFAULT 0 CHECK ("age" >= 0), description TEXT, PRIMARY KEY ("id"));`)).Parse()
		require.NoError(t, err)

		expected := `-- -
-- +CONSTRAINT users_age_check CHECK ("age" >= 0)
ALTER TABLE "users" ADD CONSTRAINT users_age_check CHECK ("age" >= 0) NOT VALID;
`
		actual, err := DiffCreateTable(
			beforeDDL.Stmts[0].(*CreateTableStmt),
			afterDDL.Stmts[0].(*CreateTableStmt),
			DiffCreateTableUseAlterTableAddConstraintNotValid(true),
		)

		assert.NoError(t, err)
		assert.Equal(t, expected, actual.String())

		t.Logf("✅: %s: actual: %%#v: \n%#v", t.Name(), actual)
		t.Logf("✅: %s: actual: %%s: \n%s", t.Name(), actual)
	})

	t.Run("success,CREATE_TABLE", func(t *testing.T) {
		t.Parallel()

		afterDDL, err := NewParser(NewLexer(`CREATE TABLE "users" (id VARCHAR(36) NOT NULL, group_id VARCHAR(36) NOT NULL REFERENCES "groups" ("id"), "name" VARCHAR(255) NOT NULL UNIQUE, "age" INT DEFAULT 0 CHECK ("age" >= 0), description TEXT, PRIMARY KEY ("id"));`)).Parse()
		require.NoError(t, err)

		expected := `CREATE TABLE "users" (
    id VARCHAR(36) NOT NULL,
    group_id VARCHAR(36) NOT NULL,
    "name" VARCHAR(255) NOT NULL,
    "age" INT DEFAULT 0,
    description TEXT,
    PRIMARY KEY ("id"),
    CONSTRAINT users_group_id_fkey FOREIGN KEY (group_id) REFERENCES "groups" ("id"),
    UNIQUE KEY users_unique_name ("name"),
    CONSTRAINT users_age_check CHECK ("age" >= 0)
);
`
		actual, err := DiffCreateTable(
			nil,
			afterDDL.Stmts[0].(*CreateTableStmt),
			DiffCreateTableUseAlterTableAddConstraintNotValid(true),
		)

		assert.NoError(t, err)
		assert.Equal(t, expected, actual.String())

		t.Logf("✅: %s: actual: %%#v: \n%#v", t.Name(), actual)
		t.Logf("✅: %s: actual: %%s: \n%s", t.Name(), actual)
	})

	t.Run("success,DROP_TABLE", func(t *testing.T) {
		t.Parallel()

		before := `CREATE TABLE "users" (id VARCHAR(36) NOT NULL, group_id VARCHAR(36) NOT NULL REFERENCES "groups" ("id"), "name" VARCHAR(255) NOT NULL UNIQUE, "age" INT DEFAULT 0 CHECK ("age" >= 0), description TEXT, PRIMARY KEY ("id"));`

		beforeDDL, err := NewParser(NewLexer(before)).Parse()
		require.NoError(t, err)

		ddls, err := DiffCreateTable(
			beforeDDL.Stmts[0].(*CreateTableStmt),
			nil,
			DiffCreateTableUseAlterTableAddConstraintNotValid(true),
		)

		assert.NoError(t, err)
		assert.Equal(t, &DDL{
			Stmts: []Stmt{
				&DropTableStmt{
					Name: &ObjectName{Name: &Ident{Name: "users", QuotationMark: `"`, Raw: `"users"`}},
				},
			},
		}, ddls)

		t.Logf("✅: %s:\n%s", t.Name(), ddls)
	})

	t.Run("success,NoAsc", func(t *testing.T) {
		t.Parallel()

		beforeDDL, err := NewParser(NewLexer(`CREATE TABLE "users" (id VARCHAR(36) NOT NULL, PRIMARY KEY ("id"));`)).Parse()
		require.NoError(t, err)

		afterDDL, err := NewParser(NewLexer(`CREATE TABLE "users" (id VARCHAR(36) NOT NULL, PRIMARY KEY ("id"));`)).Parse()
		require.NoError(t, err)

		actual, err := DiffCreateTable(
			beforeDDL.Stmts[0].(*CreateTableStmt),
			afterDDL.Stmts[0].(*CreateTableStmt),
			DiffCreateTableUseAlterTableAddConstraintNotValid(false),
		)

		assert.ErrorIs(t, err, ddl.ErrNoDifference)
		assert.Nil(t, actual)

		t.Logf("✅: %s: actual: %%#v: \n%#v", t.Name(), actual)
		t.Logf("✅: %s: actual: %%s: \n%s", t.Name(), actual)
	})
}
