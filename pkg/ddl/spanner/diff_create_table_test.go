package spanner

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

		before := `CREATE TABLE "users" (id STRING(36) NOT NULL, group_id STRING(36) NOT NULL REFERENCES "groups" ("id"), "name" STRING(255) NOT NULL UNIQUE, description STRING, PRIMARY KEY ("id"));`
		beforeDDL, err := NewParser(NewLexer(before)).Parse()
		require.NoError(t, err)

		after := `CREATE TABLE "users" (id STRING(36) NOT NULL, group_id STRING(36) NOT NULL REFERENCES "groups" ("id"), "name" STRING(255) NOT NULL UNIQUE, description STRING, PRIMARY KEY ("id"));`

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

		before := `CREATE TABLE "users" (id STRING(36) NOT NULL, group_id STRING(36) NOT NULL REFERENCES "groups" ("id"), "name" STRING(255) NOT NULL UNIQUE, description STRING, PRIMARY KEY ("id"));`
		beforeDDL, err := NewParser(NewLexer(before)).Parse()
		require.NoError(t, err)

		after := `CREATE TABLE "users" (id STRING(36) NOT NULL, group_id STRING(36) NOT NULL REFERENCES "groups" ("id"), "name" STRING(255) NOT NULL UNIQUE, "age" INT64 DEFAULT 0 NOT NULL CHECK ("age" >= 0), description STRING, PRIMARY KEY ("id"));`

		afterDDL, err := NewParser(NewLexer(after)).Parse()
		require.NoError(t, err)

		actual, err := DiffCreateTable(
			beforeDDL.Stmts[0].(*CreateTableStmt),
			afterDDL.Stmts[0].(*CreateTableStmt),
			DiffCreateTableUseAlterTableAddConstraintNotValid(false),
		)

		expectedStr := `-- -
-- +"age" INT64 NOT NULL DEFAULT 0
ALTER TABLE "users" ADD COLUMN "age" INT64 NOT NULL DEFAULT 0;
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

		before := `CREATE TABLE "users" (id STRING(36) NOT NULL, group_id STRING(36) NOT NULL REFERENCES "groups" ("id"), "name" STRING(255) NOT NULL UNIQUE, "age" INT64 DEFAULT 0 NOT NULL CHECK ("age" >= 0), description STRING, PRIMARY KEY ("id"));`
		beforeDDL, err := NewParser(NewLexer(before)).Parse()
		require.NoError(t, err)

		after := `CREATE TABLE "users" (id STRING(36) NOT NULL, group_id STRING(36) NOT NULL REFERENCES "groups" ("id"), "name" STRING(255) NOT NULL, description STRING, PRIMARY KEY ("id"));`

		afterDDL, err := NewParser(NewLexer(after)).Parse()
		require.NoError(t, err)

		actual, err := DiffCreateTable(
			beforeDDL.Stmts[0].(*CreateTableStmt),
			afterDDL.Stmts[0].(*CreateTableStmt),
			DiffCreateTableUseAlterTableAddConstraintNotValid(false),
		)

		expectedStr := `-- -UNIQUE INDEX users_unique_name (name ASC)
-- +
DROP INDEX users_unique_name;
-- -CONSTRAINT users_age_check CHECK ("age" >= 0)
-- +
ALTER TABLE "users" DROP CONSTRAINT users_age_check;
-- -"age" INT64 NOT NULL DEFAULT 0
-- +
ALTER TABLE "users" DROP COLUMN "age";
`

		assert.NoError(t, err)
		assert.Equal(t, expectedStr, actual.String())

		t.Logf("✅: %s: actual: %%#v:\n%#v", t.Name(), actual)
	})

	t.Run("success,ALTER_COLUMN_SET_DATA_TYPE", func(t *testing.T) {
		t.Parallel()

		before := `CREATE TABLE "users" (id STRING(36) NOT NULL, group_id STRING(36) NOT NULL REFERENCES "groups" ("id"), "name" STRING(255) NOT NULL, "age" INT64 DEFAULT 0 CHECK ("age" >= 0), description STRING, PRIMARY KEY ("id"));`
		beforeDDL, err := NewParser(NewLexer(before)).Parse()
		require.NoError(t, err)

		after := `CREATE TABLE "users" (id STRING(36) NOT NULL, group_id STRING(36) NOT NULL REFERENCES "groups" ("id"), "name" STRING NOT NULL UNIQUE, "age" INT64 DEFAULT 0 CHECK ("age" >= 0), description STRING, PRIMARY KEY ("id"));`

		afterDDL, err := NewParser(NewLexer(after)).Parse()
		require.NoError(t, err)

		actual, err := DiffCreateTable(
			beforeDDL.Stmts[0].(*CreateTableStmt),
			afterDDL.Stmts[0].(*CreateTableStmt),
			DiffCreateTableUseAlterTableAddConstraintNotValid(false),
		)

		expectedStr := `-- -"name" STRING(255) NOT NULL
-- +"name" STRING NOT NULL
ALTER TABLE "users" ALTER COLUMN "name" SET DATA TYPE STRING;
-- -
-- +UNIQUE INDEX users_unique_name (name ASC)
CREATE UNIQUE INDEX users_unique_name ON "users" ("name");
`

		assert.NoError(t, err)
		assert.Equal(t, expectedStr, actual.String())

		t.Logf("✅: %s: actual: %%#v:\n%#v", t.Name(), actual)
	})

	t.Run("success,ALTER_COLUMN_DROP_DEFAULT", func(t *testing.T) {
		before := `CREATE TABLE "users" (id STRING(36) NOT NULL, group_id STRING(36) NOT NULL REFERENCES "groups" ("id"), "name" STRING(255) NOT NULL UNIQUE, "age" INT64 DEFAULT 0 CHECK ("age" >= 0), description STRING, PRIMARY KEY ("id"));`
		beforeDDL, err := NewParser(NewLexer(before)).Parse()
		require.NoError(t, err)

		after := `CREATE TABLE "users" (id STRING(36) NOT NULL, group_id STRING(36) NOT NULL REFERENCES "groups" ("id"), "name" STRING(255) NOT NULL UNIQUE, "age" INT64 CHECK ("age" >= 0), description STRING, PRIMARY KEY ("id"));`
		afterDDL, err := NewParser(NewLexer(after)).Parse()
		require.NoError(t, err)

		expectedStr := `-- -"age" INT64 DEFAULT 0
-- +"age" INT64
ALTER TABLE "users" ALTER COLUMN "age" DROP DEFAULT;
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
		before := `CREATE TABLE "users" (id STRING(36) NOT NULL, group_id STRING(36) NOT NULL REFERENCES "groups" ("id"), "name" STRING(255) NOT NULL UNIQUE, "age" INT64 CHECK ("age" >= 0), description STRING, PRIMARY KEY ("id"));`
		beforeDDL, err := NewParser(NewLexer(before)).Parse()
		require.NoError(t, err)

		after := `CREATE TABLE "users" (id STRING(36) NOT NULL, group_id STRING(36) NOT NULL REFERENCES "groups" ("id"), "name" STRING(255) NOT NULL UNIQUE, "age" INT64 DEFAULT 0 CHECK ("age" <> 0), description STRING, PRIMARY KEY (id));`
		afterDDL, err := NewParser(NewLexer(after)).Parse()
		require.NoError(t, err)

		expectedStr := `-- -"age" INT64
-- +"age" INT64 DEFAULT 0
ALTER TABLE "users" ALTER COLUMN "age" SET DEFAULT 0;
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

		before := `CREATE TABLE "public.users" (id STRING(36) NOT NULL, group_id STRING(36) NOT NULL REFERENCES "groups" ("id"), "name" STRING(255) NOT NULL UNIQUE, "age" INT64 DEFAULT 0 CHECK ("age" >= 0), description STRING, PRIMARY KEY ("id"));`
		beforeDDL, err := NewParser(NewLexer(before)).Parse()
		require.NoError(t, err)

		after := `CREATE TABLE "app_users" (id STRING(36) NOT NULL, group_id STRING(36) NOT NULL REFERENCES "groups" ("id"), "name" STRING(255) NOT NULL UNIQUE, "age" INT64 DEFAULT 0 CHECK ("age" >= 0), description STRING, PRIMARY KEY ("id"));`
		afterDDL, err := NewParser(NewLexer(after)).Parse()
		require.NoError(t, err)

		expectedStr := `-- -public.users
-- +public.app_users
ALTER TABLE "public.users" RENAME TO "public.app_users";
-- -CONSTRAINT users_pkey PRIMARY KEY ("id")
-- +
ALTER TABLE "public.app_users" DROP CONSTRAINT users_pkey;
-- -CONSTRAINT users_group_id_fkey FOREIGN KEY (group_id) REFERENCES "groups" ("id")
-- +
ALTER TABLE "public.app_users" DROP CONSTRAINT users_group_id_fkey;
-- -UNIQUE INDEX users_unique_name (name ASC)
-- +
DROP INDEX public.users_unique_name;
-- -CONSTRAINT users_age_check CHECK ("age" >= 0)
-- +
ALTER TABLE "public.app_users" DROP CONSTRAINT users_age_check;
-- -
-- +CONSTRAINT app_users_pkey PRIMARY KEY ("id")
ALTER TABLE "public.app_users" ADD CONSTRAINT app_users_pkey PRIMARY KEY ("id");
-- -
-- +CONSTRAINT app_users_group_id_fkey FOREIGN KEY (group_id) REFERENCES "groups" ("id")
ALTER TABLE "public.app_users" ADD CONSTRAINT app_users_group_id_fkey FOREIGN KEY (group_id) REFERENCES "groups" ("id");
-- -
-- +UNIQUE INDEX app_users_unique_name (name ASC)
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

		before := `CREATE TABLE "users" (id STRING(36) NOT NULL, group_id STRING(36) NOT NULL REFERENCES "groups" ("id"), "name" STRING(255) NOT NULL UNIQUE, "age" INT64 DEFAULT 0 CHECK ("age" >= 0), description STRING, PRIMARY KEY ("id"));`
		beforeDDL, err := NewParser(NewLexer(before)).Parse()
		require.NoError(t, err)

		after := `CREATE TABLE "users" (id STRING(36) NOT NULL, group_id STRING(36) NOT NULL REFERENCES "groups" ("id"), "name" STRING(255) NOT NULL UNIQUE, "age" INT64 DEFAULT 0 NOT NULL CHECK ("age" >= 0), description STRING, PRIMARY KEY ("id"));`
		afterDDL, err := NewParser(NewLexer(after)).Parse()
		require.NoError(t, err)

		expectedStr := `-- -"age" INT64 DEFAULT 0
-- +"age" INT64 NOT NULL DEFAULT 0
ALTER TABLE "users" ALTER COLUMN "age" SET NOT NULL;
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

		before := `CREATE TABLE "users" (id STRING(36) NOT NULL, group_id STRING(36) NOT NULL REFERENCES "groups" ("id"), "name" STRING(255) NOT NULL UNIQUE, "age" INT64 DEFAULT 0 NOT NULL CHECK ("age" >= 0), description STRING, PRIMARY KEY ("id"));`
		beforeDDL, err := NewParser(NewLexer(before)).Parse()
		require.NoError(t, err)

		after := `CREATE TABLE "users" (id STRING(36) NOT NULL, group_id STRING(36) NOT NULL REFERENCES "groups" ("id"), "name" STRING(255) NOT NULL UNIQUE, "age" INT64 DEFAULT 0 CHECK ("age" >= 0), description STRING, PRIMARY KEY ("id"));`
		afterDDL, err := NewParser(NewLexer(after)).Parse()
		require.NoError(t, err)

		expectedStr := `-- -"age" INT64 NOT NULL DEFAULT 0
-- +"age" INT64 DEFAULT 0
ALTER TABLE "users" ALTER COLUMN "age" DROP NOT NULL;
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

		before := `CREATE TABLE "users" (id STRING(36) NOT NULL, group_id STRING(36) NOT NULL REFERENCES "groups" ("id"), "name" STRING(255) NOT NULL UNIQUE, "age" INT64 DEFAULT 0 NOT NULL CHECK ("age" >= 0), description STRING, PRIMARY KEY ("id"));`
		beforeDDL, err := NewParser(NewLexer(before)).Parse()
		require.NoError(t, err)

		after := `CREATE TABLE "users" (id STRING(36) NOT NULL, group_id STRING(36) NOT NULL REFERENCES "groups" ("id"), "name" STRING(255) NOT NULL UNIQUE, "age" INT64 DEFAULT 0 NOT NULL CHECK ("age" >= 0), description STRING, PRIMARY KEY ("id", name));`
		afterDDL, err := NewParser(NewLexer(after)).Parse()
		require.NoError(t, err)

		expectedStr := `-- -CONSTRAINT users_pkey PRIMARY KEY ("id")
-- +
ALTER TABLE "users" DROP CONSTRAINT users_pkey;
-- -
-- +CONSTRAINT users_pkey PRIMARY KEY ("id", name)
ALTER TABLE "users" ADD CONSTRAINT users_pkey PRIMARY KEY ("id", name);
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

		before := `CREATE TABLE "users" (id STRING(36) NOT NULL, group_id STRING(36) NOT NULL, "name" STRING(255) NOT NULL UNIQUE, "age" INT64 DEFAULT 0 NOT NULL CHECK ("age" >= 0), description STRING, PRIMARY KEY ("id"), CONSTRAINT users_group_id_fkey FOREIGN KEY (group_id) REFERENCES "groups" ("id"));`
		beforeDDL, err := NewParser(NewLexer(before)).Parse()
		require.NoError(t, err)

		after := `CREATE TABLE "users" (id STRING(36) NOT NULL, group_id STRING(36) NOT NULL, "name" STRING(255) NOT NULL UNIQUE, "age" INT64 DEFAULT 0 NOT NULL CHECK ("age" >= 0), description STRING, PRIMARY KEY ("id"), CONSTRAINT users_group_id_fkey FOREIGN KEY (group_id, name) REFERENCES "groups" ("id", name));`
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

		before := `CREATE TABLE "users" (id STRING(36) NOT NULL, group_id STRING(36) NOT NULL, "name" STRING(255) NOT NULL UNIQUE, "age" INT64 DEFAULT 0 NOT NULL CHECK ("age" >= 0), description STRING, PRIMARY KEY ("id"));`
		beforeDDL, err := NewParser(NewLexer(before)).Parse()
		require.NoError(t, err)

		after := `CREATE TABLE "users" (id STRING(36) NOT NULL, group_id STRING(36) NOT NULL, "name" STRING(255) NOT NULL UNIQUE, "age" INT64 DEFAULT 0 NOT NULL CHECK ("age" >= 0), description STRING, PRIMARY KEY ("id"), UNIQUE INDEX users_unique_name ("id" ASC, name ASC));`
		afterDDL, err := NewParser(NewLexer(after)).Parse()
		require.NoError(t, err)

		expectedStr := `DROP INDEX users_unique_name;
CREATE UNIQUE INDEX users_unique_name ON "users" ("id" ASC, name ASC);
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

		before := `CREATE TABLE "users" (id STRING(36) NOT NULL, group_id STRING(36) NOT NULL, "name" STRING(255) NOT NULL UNIQUE, "age" INT64 DEFAULT 0 NOT NULL CHECK ("age" >= 0), description STRING, PRIMARY KEY ("id"));`
		beforeDDL, err := NewParser(NewLexer(before)).Parse()
		require.NoError(t, err)

		after := `CREATE TABLE "users" (id STRING(36) NOT NULL, group_id STRING(36) NOT NULL, "name" STRING(255) NOT NULL UNIQUE, "age" INT64 DEFAULT ( (0 + 3) - 1 * 4 / 2 ) NOT NULL CHECK ("age" >= 0), description STRING, PRIMARY KEY ("id"));`
		afterDDL, err := NewParser(NewLexer(after)).Parse()
		require.NoError(t, err)

		expectedStr := `-- -"age" INT64 NOT NULL DEFAULT 0
-- +"age" INT64 NOT NULL DEFAULT ((0 + 3) - 1 * 4 / 2)
ALTER TABLE "users" ALTER COLUMN "age" SET DEFAULT ((0 + 3) - 1 * 4 / 2);
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

	t.Run("success,ALTER_COLUMN_SET_DEFAULT_complex", func(t *testing.T) {
		t.Parallel()

		before := `CREATE TABLE complex_defaults (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    unique_code STRING,
    status STRING DEFAULT 'pending',
    random_number INT64 DEFAULT FLOOR(RANDOM() * 100)::INTEGER,
    json_data JSONB DEFAULT '{}',
    calculated_value INT64 DEFAULT (SELECT COUNT(*) FROM another_table)
);
`
		beforeDDL, err := NewParser(NewLexer(before)).Parse()
		require.NoError(t, err)

		after := `CREATE TABLE complex_defaults (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    unique_code STRING DEFAULT 'CODE-' || TO_CHAR(NOW(), 'YYYYMMDDHH24MISS') || '-' || LPAD(TO_CHAR(NEXTVAL('seq_complex_default')), 5, '0'),
    status STRING DEFAULT 'pending',
    random_number INT64 DEFAULT FLOOR(RANDOM() * 100)::INTEGER,
    json_data JSONB DEFAULT '{}',
    calculated_value INT64 DEFAULT (SELECT COUNT(*) FROM another_table)
);
`
		afterDDL, err := NewParser(NewLexer(after)).Parse()
		require.NoError(t, err)

		expectedStr := `-- -unique_code STRING
-- +unique_code STRING DEFAULT 'CODE-' || TO_CHAR(NOW(), 'YYYYMMDDHH24MISS') || '-' || LPAD(TO_CHAR(NEXTVAL('seq_complex_default')), 5, '0')
ALTER TABLE complex_defaults ALTER COLUMN unique_code SET DEFAULT 'CODE-' || TO_CHAR(NOW(), 'YYYYMMDDHH24MISS') || '-' || LPAD(TO_CHAR(NEXTVAL('seq_complex_default')), 5, '0');
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

	t.Run("success,DiffCreateTableUseAlterTableAddConstraintNotValid", func(t *testing.T) {
		t.Parallel()

		beforeDDL, err := NewParser(NewLexer(`CREATE TABLE "users" (id STRING(36) NOT NULL, group_id STRING(36) NOT NULL REFERENCES "groups" ("id"), "name" STRING(255) NOT NULL UNIQUE, "age" INT64 DEFAULT 0, description STRING, PRIMARY KEY ("id"));`)).Parse()
		require.NoError(t, err)

		afterDDL, err := NewParser(NewLexer(`CREATE TABLE "users" (id STRING(36) NOT NULL, group_id STRING(36) NOT NULL REFERENCES "groups" ("id"), "name" STRING(255) NOT NULL UNIQUE, "age" INT64 DEFAULT 0 CHECK ("age" >= 0), description STRING, PRIMARY KEY ("id"));`)).Parse()
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

		afterDDL, err := NewParser(NewLexer(`CREATE TABLE "users" (id STRING(36) NOT NULL, group_id STRING(36) NOT NULL REFERENCES "groups" ("id"), "name" STRING(255) NOT NULL UNIQUE, "age" INT64 DEFAULT 0 CHECK ("age" >= 0), description STRING, PRIMARY KEY ("id"));`)).Parse()
		require.NoError(t, err)

		expected := `CREATE TABLE "users" (
    id STRING(36) NOT NULL,
    group_id STRING(36) NOT NULL,
    "name" STRING(255) NOT NULL,
    "age" INT64 DEFAULT 0,
    description STRING,
    CONSTRAINT users_pkey PRIMARY KEY ("id"),
    CONSTRAINT users_group_id_fkey FOREIGN KEY (group_id) REFERENCES "groups" ("id"),
    UNIQUE INDEX users_unique_name ("name"),
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

		before := `CREATE TABLE "users" (id STRING(36) NOT NULL, group_id STRING(36) NOT NULL REFERENCES "groups" ("id"), "name" STRING(255) NOT NULL UNIQUE, "age" INT64 DEFAULT 0 CHECK ("age" >= 0), description STRING, PRIMARY KEY ("id"));`

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

		beforeDDL, err := NewParser(NewLexer(`CREATE TABLE "users" (id STRING(36) NOT NULL, PRIMARY KEY ("id" ASC));`)).Parse()
		require.NoError(t, err)

		afterDDL, err := NewParser(NewLexer(`CREATE TABLE "users" (id STRING(36) NOT NULL, PRIMARY KEY ("id"));`)).Parse()
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
