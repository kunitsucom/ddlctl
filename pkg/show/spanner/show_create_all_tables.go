package spanner

import (
	"context"
	"database/sql"
	"fmt"

	sqlz "github.com/kunitsucom/util.go/database/sql"
	errorz "github.com/kunitsucom/util.go/errors"

	"github.com/kunitsucom/ddlctl/pkg/internal/logs"
)

type sqlQueryerContext = interface {
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
}

const (
	querySelectTableName = `SELECT TABLE_NAME FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_SCHEMA = '';`
)

type informationSchemaTable struct {
	TableName string `db:"TABLE_NAME"`
}

const (
	queryShowCreateAllTables = `SELECT TABLE_NAME, COLUMN_NAME, COLUMN_DEFAULT, IS_NULLABLE, SPANNER_TYPE FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_NAME = ? ORDER BY TABLE_NAME, ORDINAL_POSITION;`
)

type informationSchemaColumn struct {
	TableName     string  `db:"TABLE_NAME"`
	ColumnName    string  `db:"COLUMN_NAME"`
	ColumnDefault *string `db:"COLUMN_DEFAULT"`
	IsNullable    string  `db:"IS_NULLABLE"`
	SpannerType   string  `db:"SPANNER_TYPE"`
}

func (c *informationSchemaColumn) String() string {
	d := fmt.Sprintf("%s %s", c.ColumnName, c.SpannerType)
	if c.ColumnDefault != nil {
		d += fmt.Sprintf(" DEFAULT %s", *c.ColumnDefault)
	}
	if c.IsNullable == "NO" {
		d += " NOT NULL"
	}
	return d
}

const (
	queryShowPrimaryKey = `SELECT i.INDEX_NAME, i.INDEX_TYPE, ic.COLUMN_NAME, ic.COLUMN_ORDERING, ic.ORDINAL_POSITION FROM INFORMATION_SCHEMA.INDEXES AS i INNER JOIN INFORMATION_SCHEMA.INDEX_COLUMNS AS ic ON i.TABLE_NAME = ic.TABLE_NAME WHERE i.TABLE_NAME = ? AND i.INDEX_TYPE = "PRIMARY_KEY" ORDER BY i.TABLE_NAME, ic.ORDINAL_POSITION;`
)

type informationSchemaPrimaryKey struct {
	// INDEXES https://cloud.google.com/spanner/docs/information-schema?hl=ja#indexes
	// INDEX_COLUMNS https://cloud.google.com/spanner/docs/information-schema?hl=ja#index_columns
	IndexName       string `db:"INDEX_NAME"`
	IndexType       string `db:"INDEX_TYPE"`
	ColumnName      string `db:"COLUMN_NAME"`
	ColumnOrdering  string `db:"COLUMN_ORDERING"`
	OrdinalPosition int    `db:"ORDINAL_POSITION"`
}

const (
	querySelectIndexes = `SELECT DISTINCT INDEX_NAME, IS_UNIQUE FROM INFORMATION_SCHEMA.INDEXES WHERE TABLE_NAME = ? AND INDEX_TYPE != "PRIMARY_KEY";`
)

type informationSchemaIndexName struct {
	IndexName string `db:"INDEX_NAME"`
	IsUnique  bool   `db:"IS_UNIQUE"`
}

const (
	queryShowIndexes = `SELECT ic.INDEX_NAME, i.INDEX_TYPE, ic.COLUMN_NAME, ic.COLUMN_ORDERING, ic.ORDINAL_POSITION FROM INFORMATION_SCHEMA.INDEXES AS i INNER JOIN INFORMATION_SCHEMA.INDEX_COLUMNS AS ic ON i.TABLE_NAME = ic.TABLE_NAME WHERE i.TABLE_NAME = ? AND i.INDEX_TYPE != "PRIMARY_KEY" AND ic.INDEX_NAME != "PRIMARY_KEY" ORDER BY i.TABLE_NAME, ic.INDEX_NAME, ic.ORDINAL_POSITION;`
)

type informationSchemaIndex struct {
	// INDEXES https://cloud.google.com/spanner/docs/information-schema?hl=ja#indexes
	// INDEX_COLUMNS https://cloud.google.com/spanner/docs/information-schema?hl=ja#index_columns
	IndexName       string `db:"INDEX_NAME"`
	IndexType       string `db:"INDEX_TYPE"`
	ColumnName      string `db:"COLUMN_NAME"`
	ColumnOrdering  string `db:"COLUMN_ORDERING"`
	OrdinalPosition int    `db:"ORDINAL_POSITION"`
}

type showCreateAllTablesConfig struct {
	schema string
}

type ShowCreateAllTablesOption interface {
	apply(cfg *showCreateAllTablesConfig)
}

type showCreateAllTablesOptionSchema struct{ schema string }

func (o *showCreateAllTablesOptionSchema) apply(config *showCreateAllTablesConfig) {
	config.schema = o.schema
}

func WithShowCreateAllTablesOptionSchema(schema string) ShowCreateAllTablesOption { //nolint:ireturn
	return &showCreateAllTablesOptionSchema{schema: schema}
}

//nolint:cyclop,funlen,gocognit
func ShowCreateAllTables(ctx context.Context, db sqlQueryerContext, opts ...ShowCreateAllTablesOption) (query string, err error) {
	dbz := sqlz.NewDB(db)

	cfg := &showCreateAllTablesConfig{
		schema: "public",
	}
	for _, opt := range opts {
		opt.apply(cfg)
	}

	tables := make([]*informationSchemaTable, 0)
	if err := dbz.QueryContext(ctx, &tables, querySelectTableName); err != nil {
		return "", errorz.Errorf("dbz.QueryContext: %w", err)
	}

	tablesLastIndex := len(tables) - 1
	for tblIdx, tbl := range tables {
		// TABLE
		d := fmt.Sprintf("CREATE TABLE %s (\n", tbl.TableName)

		columns := make([]*informationSchemaColumn, 0)
		if err := dbz.QueryContext(ctx, &columns, queryShowCreateAllTables, tbl.TableName); err != nil {
			return "", errorz.Errorf("dbz.QueryContext: %w", err)
		}

		columnsLastIndex := len(columns) - 1
		for i, col := range columns {
			d += fmt.Sprintf("    %s", col)
			if i != columnsLastIndex {
				d += ","
			}
			d += "\n"

			logs.Trace.Printf("table=%s: columns: %s", tbl.TableName, col)
		}
		d += ")"

		primaryKeyColumns := make([]*informationSchemaPrimaryKey, 0)
		if err := dbz.QueryContext(ctx, &primaryKeyColumns, queryShowPrimaryKey, tbl.TableName); err != nil {
			return "", errorz.Errorf("dbz.QueryContext: %w", err)
		}

		if len(primaryKeyColumns) > 0 {
			d += " PRIMARY KEY ("
			primaryKeyColumnsLastIndex := len(primaryKeyColumns) - 1
			for i, pk := range primaryKeyColumns {
				d += pk.ColumnName
				if i != primaryKeyColumnsLastIndex {
					d += ", "
				}
			}
			d += ")"
		}

		// append table
		query += d + ";\n"

		// INDEX
		indexNames := make([]*informationSchemaIndexName, 0)
		if err := dbz.QueryContext(ctx, &indexNames, querySelectIndexes, tbl.TableName); err != nil {
			return "", errorz.Errorf("dbz.QueryContext: %w", err)
		}

		for _, indexName := range indexNames {
			indexes := make([]*informationSchemaIndex, 0)
			if err := dbz.QueryContext(ctx, &indexes, queryShowIndexes, tbl.TableName); err != nil {
				return "", errorz.Errorf("dbz.QueryContext: %w", err)
			}

			d := "CREATE "
			if indexName.IsUnique {
				d += "UNIQUE "
			}
			d += fmt.Sprintf("INDEX %s ON %s (", indexName.IndexName, tbl.TableName)

			indexesLastIndex := len(indexes) - 1
			for i, idx := range indexes {
				d += idx.ColumnName
				if i != indexesLastIndex {
					d += ", "
				}
			}
			d += ");\n"

			// append index
			query += d
		}

		if tblIdx != tablesLastIndex {
			query += "\n"
		}
	}

	return query, nil
}
