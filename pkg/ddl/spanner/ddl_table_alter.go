package spanner

import (
	"strings"

	"github.com/kunitsucom/ddlctl/pkg/ddl/internal"
)

// MEMO: https://cloud.google.com/spanner/docs/reference/standard-sql/data-definition-language#alter_table

var _ Stmt = (*AlterTableStmt)(nil)

type AlterTableStmt struct {
	Comment string
	Indent  string
	Name    *ObjectName
	Action  AlterTableAction
}

func (*AlterTableStmt) isStmt() {}

func (s *AlterTableStmt) GetNameForDiff() string {
	return s.Name.StringForDiff()
}

//nolint:cyclop,funlen
func (s *AlterTableStmt) String() string {
	var str string
	if s.Comment != "" {
		comments := strings.Split(s.Comment, "\n")
		for i := range comments {
			if comments[i] != "" {
				str += CommentPrefix + comments[i] + "\n"
			}
		}
	}
	str += "ALTER TABLE "
	str += s.Name.String() + " "
	switch a := s.Action.(type) {
	case *RenameTable:
		str += "RENAME TO "
		str += a.NewName.String()
	case *RenameColumn:
		str += "RENAME COLUMN " + a.Name.String() + " TO " + a.NewName.String()
	case *RenameConstraint:
		str += "RENAME CONSTRAINT " + a.Name.String() + " TO " + a.NewName.String()
	case *AddColumn:
		str += "ADD COLUMN " + a.Column.String()
	case *DropColumn:
		str += "DROP COLUMN " + a.Name.String()
	case *AlterColumn:
		str += "ALTER COLUMN " + a.Name.String() + " "
		switch ca := a.Action.(type) {
		case *AlterColumnDataType:
			str += ca.DataType.String()
			if ca.NotNull {
				str += " NOT NULL"
			}
		case *AlterColumnSetDefault:
			str += "SET " + ca.Default.String()
		case *AlterColumnDropDefault:
			str += "DROP DEFAULT"
		case *AlterColumnSetOptions:
			str += "SET OPTIONS " + ca.Options.String()
		case *AlterColumnDropOptions:
			str += "DROP OPTIONS"
		}
	case *AddConstraint:
		str += "ADD " + a.Constraint.String()
		if a.NotValid {
			str += " NOT VALID"
		}
	case *DropConstraint:
		str += "DROP CONSTRAINT " + a.Name.String()
	case *AlterConstraint:
		str += "ALTER CONSTRAINT " + a.Name.String() + " "
		if a.Deferrable {
			str += "DEFERRABLE"
		} else {
			str += "NOT DEFERRABLE"
		}
		if a.InitiallyDeferred {
			str += " INITIALLY DEFERRED"
		} else {
			str += " INITIALLY IMMEDIATE"
		}
	}

	return str + ";\n"
}

func (s *AlterTableStmt) GoString() string { return internal.GoString(*s) }

type AlterTableAction interface {
	isAlterTableAction()
	GoString() string
}

// RenameTable represents ALTER TABLE table_name RENAME TO new_table_name.
type RenameTable struct {
	NewName *ObjectName
}

func (*RenameTable) isAlterTableAction() {}

func (s *RenameTable) GoString() string { return internal.GoString(*s) }

// RenameConstraint represents ALTER TABLE table_name RENAME COLUMN.
type RenameConstraint struct {
	Name    *Ident
	NewName *Ident
}

func (*RenameConstraint) isAlterTableAction() {}

func (s *RenameConstraint) GoString() string { return internal.GoString(*s) }

// RenameColumn represents ALTER TABLE table_name RENAME COLUMN.
type RenameColumn struct {
	Name    *Ident
	NewName *Ident
}

func (*RenameColumn) isAlterTableAction() {}

func (s *RenameColumn) GoString() string { return internal.GoString(*s) }

// AddColumn represents ALTER TABLE table_name ADD COLUMN.
type AddColumn struct {
	Column *Column
}

func (*AddColumn) isAlterTableAction() {}

func (s *AddColumn) GoString() string { return internal.GoString(*s) }

// DropColumn represents ALTER TABLE table_name DROP COLUMN.
type DropColumn struct {
	Name *Ident
}

func (*DropColumn) isAlterTableAction() {}

func (s *DropColumn) GoString() string { return internal.GoString(*s) }

// AlterColumn represents ALTER TABLE table_name ALTER COLUMN.
type AlterColumn struct {
	Name   *Ident
	Action AlterColumnAction
}

func (*AlterColumn) isAlterTableAction() {}

func (s *AlterColumn) GoString() string { return internal.GoString(*s) }

type AlterColumnAction interface {
	isAlterColumnAction()
	GoString() string
}

// AlterColumnDataType represents ALTER TABLE table_name ALTER COLUMN column_name SET DATA TYPE.
type AlterColumnDataType struct {
	DataType *DataType
	NotNull  bool
}

func (*AlterColumnDataType) isAlterColumnAction() {}

func (s *AlterColumnDataType) GoString() string { return internal.GoString(*s) }

// AlterColumnSetDefault represents ALTER TABLE table_name ALTER COLUMN column_name SET DEFAULT.
type AlterColumnSetDefault struct {
	Default *Default
}

func (*AlterColumnSetDefault) isAlterColumnAction() {}

func (s *AlterColumnSetDefault) GoString() string { return internal.GoString(*s) }

// AlterColumnDropDefault represents ALTER TABLE table_name ALTER COLUMN column_name DROP DEFAULT.
type AlterColumnDropDefault struct{}

func (*AlterColumnDropDefault) isAlterColumnAction() {}

func (s *AlterColumnDropDefault) GoString() string { return internal.GoString(*s) }

// AlterColumnSetOptions represents ALTER TABLE table_name ALTER COLUMN column_name SET OPTIONS.
type AlterColumnSetOptions struct {
	Options *Expr
}

func (*AlterColumnSetOptions) isAlterColumnAction() {}

func (s *AlterColumnSetOptions) GoString() string { return internal.GoString(*s) }

// AlterColumnDropOptions represents ALTER TABLE table_name ALTER COLUMN column_name DROP OPTIONS.
type AlterColumnDropOptions struct{}

func (*AlterColumnDropOptions) isAlterColumnAction() {}

func (s *AlterColumnDropOptions) GoString() string { return internal.GoString(*s) }

// AddConstraint represents ALTER TABLE table_name ADD CONSTRAINT.
type AddConstraint struct {
	Constraint Constraint
	NotValid   bool
}

func (*AddConstraint) isAlterTableAction() {}

func (s *AddConstraint) GoString() string { return internal.GoString(*s) }

// DropConstraint represents ALTER TABLE table_name DROP CONSTRAINT.
type DropConstraint struct {
	Name *Ident
}

func (*DropConstraint) isAlterTableAction() {}

func (s *DropConstraint) GoString() string { return internal.GoString(*s) }

// AlterConstraint represents ALTER TABLE table_name ALTER CONSTRAINT.
type AlterConstraint struct {
	Name              *Ident
	Deferrable        bool
	InitiallyDeferred bool
}

func (*AlterConstraint) isAlterTableAction() {}

func (s *AlterConstraint) GoString() string { return internal.GoString(*s) }
