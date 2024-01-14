package mysql

// MEMO: https://dev.mysql.com/doc/refman/8.0/en/create-table-check-constraints.html

import (
	"fmt"
	"runtime"
	"strings"

	filepathz "github.com/kunitsucom/util.go/path/filepath"
	stringz "github.com/kunitsucom/util.go/strings"

	apperr "github.com/kunitsucom/ddlctl/pkg/apperr"

	"github.com/kunitsucom/ddlctl/pkg/ddl"
	"github.com/kunitsucom/ddlctl/pkg/ddl/logs"
)

//nolint:gochecknoglobals
var quotationMarks = []string{`"`, "`", "'"}

func NewRawIdent(raw string) *Ident {
	for _, q := range quotationMarks {
		if strings.HasPrefix(raw, q) && strings.HasSuffix(raw, q) {
			return &Ident{
				Name:          strings.Trim(raw, q),
				QuotationMark: q,
				Raw:           raw,
			}
		}
	}

	return &Ident{
		Name:          raw,
		QuotationMark: "",
		Raw:           raw,
	}
}

func NewIdent(name, quotationMark, raw string) *Ident {
	return &Ident{
		Name:          name,
		QuotationMark: quotationMark,
		Raw:           raw,
	}
}

// Parser はSQL文を解析するパーサーです。
type Parser struct {
	l            *Lexer
	currentToken Token
	peekToken    Token
}

// NewParser は新しいParserを生成します。
func NewParser(l *Lexer) *Parser {
	p := &Parser{
		l: l,
	}

	return p
}

// nextToken は次のトークンを読み込みます。
func (p *Parser) nextToken() {
	p.currentToken = p.peekToken
	p.peekToken = p.l.NextToken()

	_, file, line, _ := runtime.Caller(1)
	logs.TraceLog.Printf("🪲: nextToken: caller=%s:%d currentToken: %#v, peekToken: %#v", filepathz.Short(file), line, p.currentToken, p.peekToken)
}

// Parse はSQL文を解析します。
func (p *Parser) Parse() (*DDL, error) { //nolint:ireturn
	p.nextToken() // current = ""
	p.nextToken() // current = CREATE or ALTER or ...

	d := &DDL{}

LabelDDL:
	for {
		switch p.currentToken.Type { //nolint:exhaustive
		case TOKEN_CREATE:
			stmt, err := p.parseCreateStatement()
			if err != nil {
				return nil, apperr.Errorf("parseCreateStatement: %w", err)
			}
			d.Stmts = append(d.Stmts, stmt)
		case TOKEN_CLOSE_PAREN:
			// do nothing
		case TOKEN_SEMICOLON:
			// do nothing
		case TOKEN_EOF:
			break LabelDDL
		default:
			return nil, apperr.Errorf("currentToken=%#v: %w", p.currentToken, ddl.ErrUnexpectedToken)
		}

		p.nextToken()
	}
	return d, nil
}

func (p *Parser) parseCreateStatement() (Stmt, error) { //nolint:ireturn
	p.nextToken() // current = TABLE or INDEX or ...

	switch p.currentToken.Type { //nolint:exhaustive
	case TOKEN_TABLE:
		return p.parseCreateTableStmt()
	case TOKEN_INDEX, TOKEN_UNIQUE:
		return p.parseCreateIndexStmt()
	default:
		return nil, apperr.Errorf("currentToken=%#v: %w", p.currentToken, ddl.ErrUnexpectedToken)
	}
}

//nolint:cyclop,funlen,gocognit
func (p *Parser) parseCreateTableStmt() (*CreateTableStmt, error) {
	createTableStmt := &CreateTableStmt{
		Indent: Indent,
	}

	if p.isPeekToken(TOKEN_IF) {
		p.nextToken() // current = IF
		if err := p.checkPeekToken(TOKEN_NOT); err != nil {
			return nil, apperr.Errorf("checkPeekToken: %w", err)
		}
		p.nextToken() // current = NOT
		if err := p.checkPeekToken(TOKEN_EXISTS); err != nil {
			return nil, apperr.Errorf("checkPeekToken: %w", err)
		}
		p.nextToken() // current = EXISTS
		createTableStmt.IfNotExists = true
	}

	p.nextToken() // current = table_name
	if err := p.checkCurrentToken(TOKEN_IDENT); err != nil {
		return nil, apperr.Errorf("checkCurrentToken: %w", err)
	}

	createTableStmt.Name = NewObjectName(p.currentToken.Literal.Str)
	errFmtPrefix := fmt.Sprintf("table_name=%s: ", createTableStmt.Name.StringForDiff())

	p.nextToken() // current = (

	if err := p.checkCurrentToken(TOKEN_OPEN_PAREN); err != nil {
		return nil, apperr.Errorf(errFmtPrefix+"checkCurrentToken: %w", err)
	}

	p.nextToken() // current = column_name

LabelColumns:
	for {
		switch { //nolint:exhaustive
		case p.isCurrentToken(TOKEN_IDENT):
			column, constraints, err := p.parseColumn(createTableStmt.Name.Name)
			if err != nil {
				return nil, apperr.Errorf(errFmtPrefix+"parseColumn: %w", err)
			}
			createTableStmt.Columns = append(createTableStmt.Columns, column)
			if len(constraints) > 0 {
				for _, c := range constraints {
					createTableStmt.Constraints = createTableStmt.Constraints.Append(c)
				}
			}
		case isConstraint(p.currentToken.Type):
			constraint, err := p.parseTableConstraint(createTableStmt.Name.Name)
			if err != nil {
				return nil, apperr.Errorf(errFmtPrefix+"parseConstraint: %w", err)
			}
			createTableStmt.Constraints = createTableStmt.Constraints.Append(constraint)
		case p.isCurrentToken(TOKEN_COMMA):
			p.nextToken()
			continue
		case p.isCurrentToken(TOKEN_CLOSE_PAREN):
			p.nextToken()
			break LabelColumns
		default:
			return nil, apperr.Errorf(errFmtPrefix+"currentToken=%#v: %w", p.currentToken, ddl.ErrUnexpectedToken)
		}
	}

LabelTableOptions:
	for {
		opt := &Option{}
		switch p.currentToken.Type { //nolint:exhaustive
		case TOKEN_ENGINE:
			opt.Name = "ENGINE"
			p.nextToken() // current = `=`
			if err := p.checkCurrentToken(TOKEN_EQUAL); err != nil {
				return nil, apperr.Errorf(errFmtPrefix+"checkCurrentToken: %w", err)
			}
			p.nextToken() // current = TOKEN_IDENT
			if err := p.checkCurrentToken(TOKEN_IDENT); err != nil {
				return nil, apperr.Errorf(errFmtPrefix+"checkCurrentToken: %w", err)
			}
			opt.Value = NewRawIdent(p.currentToken.Literal.Str)
		case TOKEN_DEFAULT:
			if err := p.checkPeekToken(TOKEN_CHARSET); err != nil {
				return nil, apperr.Errorf(errFmtPrefix+"checkPeekToken: %w", err)
			}
			p.nextToken() // current = CHARSET
			opt.Name = "DEFAULT CHARSET"
			p.nextToken() // current = `=`
			if err := p.checkCurrentToken(TOKEN_EQUAL); err != nil {
				return nil, apperr.Errorf(errFmtPrefix+"checkCurrentToken: %w", err)
			}
			p.nextToken() // current = TOKEN_IDENT
			if err := p.checkCurrentToken(TOKEN_IDENT); err != nil {
				return nil, apperr.Errorf(errFmtPrefix+"checkCurrentToken: %w", err)
			}
			opt.Value = NewRawIdent(p.currentToken.Literal.Str)
		case TOKEN_COLLATE:
			opt.Name = "COLLATE"
			p.nextToken() // current = `=`
			if err := p.checkCurrentToken(TOKEN_EQUAL); err != nil {
				return nil, apperr.Errorf(errFmtPrefix+"checkCurrentToken: %w", err)
			}
			p.nextToken() // current = TOKEN_IDENT
			if err := p.checkCurrentToken(TOKEN_IDENT); err != nil {
				return nil, apperr.Errorf(errFmtPrefix+"checkCurrentToken: %w", err)
			}
			opt.Value = NewRawIdent(p.currentToken.Literal.Str)
		case TOKEN_SEMICOLON, TOKEN_EOF:
			break LabelTableOptions
		default:
			return nil, apperr.Errorf(errFmtPrefix+"peekToken=%#v: %w", p.peekToken, ddl.ErrUnexpectedToken)
		}
		createTableStmt.Options = append(createTableStmt.Options, opt)
		p.nextToken()
	}

	return createTableStmt, nil
}

//nolint:cyclop,funlen
func (p *Parser) parseCreateIndexStmt() (*CreateIndexStmt, error) {
	createIndexStmt := &CreateIndexStmt{}

	if p.isCurrentToken(TOKEN_UNIQUE) {
		createIndexStmt.Unique = true
		p.nextToken() // current = INDEX
	}

	if p.isPeekToken(TOKEN_IF) {
		p.nextToken() // current = IF
		if err := p.checkPeekToken(TOKEN_NOT); err != nil {
			return nil, apperr.Errorf("checkPeekToken: %w", err)
		}
		p.nextToken() // current = NOT
		if err := p.checkPeekToken(TOKEN_EXISTS); err != nil {
			return nil, apperr.Errorf("checkPeekToken: %w", err)
		}
		p.nextToken() // current = EXISTS
		createIndexStmt.IfNotExists = true
	}

	p.nextToken() // current = index_name
	if err := p.checkCurrentToken(TOKEN_IDENT); err != nil {
		return nil, apperr.Errorf("checkCurrentToken: %w", err)
	}

	createIndexStmt.Name = NewObjectName(p.currentToken.Literal.Str)
	errFmtPrefix := fmt.Sprintf("index_name=%s: ", createIndexStmt.Name.StringForDiff())

	p.nextToken() // current = ON

	if err := p.checkCurrentToken(TOKEN_ON); err != nil {
		return nil, apperr.Errorf(errFmtPrefix+"checkCurrentToken: %w", err)
	}

	p.nextToken() // current = table_name

	if err := p.checkCurrentToken(TOKEN_IDENT); err != nil {
		return nil, apperr.Errorf(errFmtPrefix+"checkCurrentToken: %w", err)
	}

	createIndexStmt.TableName = NewObjectName(p.currentToken.Literal.Str)

	p.nextToken() // current = USING or (

	if p.isCurrentToken(TOKEN_USING) {
		p.nextToken() // current = using_def
		createIndexStmt.Using = append(createIndexStmt.Using, NewIdent(p.currentToken.Literal.Str, "", p.currentToken.Literal.Str))
		p.nextToken() // current = (
	}

	if err := p.checkCurrentToken(TOKEN_OPEN_PAREN); err != nil {
		return nil, apperr.Errorf(errFmtPrefix+"checkCurrentToken: %w", err)
	}

	idents, err := p.parseColumnIdents()
	if err != nil {
		return nil, apperr.Errorf(errFmtPrefix+"parseColumnIdents: %w", err)
	}

	createIndexStmt.Columns = idents

	return createIndexStmt, nil
}

//nolint:funlen,cyclop,gocognit
func (p *Parser) parseColumn(tableName *Ident) (*Column, []Constraint, error) {
	column := &Column{}
	constraints := make(Constraints, 0)

	if err := p.checkCurrentToken(TOKEN_IDENT); err != nil {
		return nil, nil, apperr.Errorf("checkCurrentToken: %w", err)
	}

	column.Name = NewRawIdent(p.currentToken.Literal.Str)
	errFmtPrefix := fmt.Sprintf("column_name=%s: ", column.Name.StringForDiff())

	p.nextToken() // current = DATA_TYPE

	switch { //nolint:exhaustive
	case isDataType(p.currentToken.Type):
		dataType, err := p.parseDataType()
		if err != nil {
			return nil, nil, apperr.Errorf(errFmtPrefix+"parseDataType: %w", err)
		}
		column.DataType = dataType

		p.nextToken() // current = DEFAULT or NOT or NULL or PRIMARY or UNIQUE or COMMA or ...
	LabelDefaultNotNull:
		for {
			switch p.currentToken.Type { //nolint:exhaustive
			case TOKEN_NOT:
				if err := p.checkPeekToken(TOKEN_NULL); err != nil {
					return nil, nil, apperr.Errorf(errFmtPrefix+"checkPeekToken: %w", err)
				}
				p.nextToken() // current = NULL
				column.NotNull = true
			case TOKEN_NULL:
				column.NotNull = false
			case TOKEN_DEFAULT:
				p.nextToken() // current = default_value
				def, err := p.parseColumnDefault()
				if err != nil {
					return nil, nil, apperr.Errorf(errFmtPrefix+"parseColumnDefault: %w", err)
				}
				column.Default = def
				continue
			default:
				break LabelDefaultNotNull
			}

			p.nextToken()
		}

		cs, err := p.parseColumnConstraints(tableName, column)
		if err != nil {
			return nil, nil, apperr.Errorf(errFmtPrefix+"parseColumnConstraints: %w", err)
		}
		if len(cs) > 0 {
			for _, c := range cs {
				constraints = constraints.Append(c)
			}
		}
	default:
		return nil, nil, apperr.Errorf(errFmtPrefix+"currentToken=%#v: %w", p.currentToken, ddl.ErrUnexpectedToken)
	}

	return column, constraints, nil
}

//nolint:cyclop
func (p *Parser) parseColumnDefault() (*Default, error) {
	def := &Default{}

LabelDefault:
	for {
		switch p.currentToken.Type { //nolint:exhaustive
		case TOKEN_IDENT:
			def.Value = def.Value.Append(NewRawIdent(p.currentToken.Literal.String()))
		case TOKEN_OPEN_PAREN:
			ids, err := p.parseExpr()
			if err != nil {
				return nil, apperr.Errorf("parseExpr: %w", err)
			}
			def.Value = def.Value.Append(ids...)
			continue
		case TOKEN_NOT, TOKEN_NULL, TOKEN_COMMA, TOKEN_CLOSE_PAREN:
			break LabelDefault
		default:
			if isReservedValue(p.currentToken.Type) {
				def.Value = def.Value.Append(NewIdent(string(p.currentToken.Type), "", p.currentToken.Literal.String()))
				p.nextToken()
				continue
			}
			// MEMO: backup
			// TODO: check if this is necessary
			// if isOperator(p.currentToken.Type) {
			// 	def.Value = def.Value.Append(NewRawIdent(p.currentToken.Literal.Str))
			// 	p.nextToken()
			// 	continue
			// }
			// if isDataType(p.currentToken.Type) {
			// 	def.Value.Idents = append(def.Value.Idents, NewRawIdent(p.currentToken.Literal.Str))
			// 	p.nextToken()
			// 	continue
			// }
			if isConstraint(p.currentToken.Type) {
				break LabelDefault
			}
			return nil, apperr.Errorf("currentToken=%#v: %w", p.currentToken, ddl.ErrUnexpectedToken)
		}

		p.nextToken()
	}

	return def, nil
}

func (p *Parser) parseExpr() ([]*Ident, error) {
	idents := make([]*Ident, 0)

	if err := p.checkCurrentToken(TOKEN_OPEN_PAREN); err != nil {
		return nil, apperr.Errorf("checkCurrentToken: %w", err)
	}
	idents = append(idents, NewRawIdent(p.currentToken.Literal.Str))
	p.nextToken() // current = IDENT

LabelExpr:
	for {
		switch p.currentToken.Type { //nolint:exhaustive
		case TOKEN_OPEN_PAREN:
			ids, err := p.parseExpr()
			if err != nil {
				return nil, apperr.Errorf("parseExpr: %w", err)
			}
			idents = append(idents, ids...)
			continue
		case TOKEN_CLOSE_PAREN:
			idents = append(idents, NewRawIdent(p.currentToken.Literal.Str))
			p.nextToken()
			break LabelExpr
		case TOKEN_EQUAL, TOKEN_GREATER, TOKEN_LESS:
			value := p.currentToken.Literal.Str
			switch p.peekToken.Type { //nolint:exhaustive
			case TOKEN_EQUAL, TOKEN_GREATER, TOKEN_LESS:
				value += p.peekToken.Literal.Str
				p.nextToken()
			}
			idents = append(idents, NewRawIdent(value))
		case TOKEN_EOF:
			return nil, apperr.Errorf("currentToken=%#v: %w", p.currentToken, ddl.ErrUnexpectedToken)
		default:
			if isReservedValue(p.currentToken.Type) {
				idents = append(idents, NewRawIdent(p.currentToken.Type.String()))
			} else {
				idents = append(idents, NewRawIdent(p.currentToken.Literal.Str))
			}
		}

		p.nextToken()
	}

	return idents, nil
}

//nolint:cyclop,funlen,gocognit
func (p *Parser) parseColumnConstraints(tableName *Ident, column *Column) ([]Constraint, error) {
	constraints := make(Constraints, 0)

LabelConstraints:
	for {
		switch p.currentToken.Type { //nolint:exhaustive
		case TOKEN_PRIMARY:
			if err := p.checkPeekToken(TOKEN_KEY); err != nil {
				return nil, apperr.Errorf("checkPeekToken: %w", err)
			}
			p.nextToken() // current = KEY
			constraints = constraints.Append(&PrimaryKeyConstraint{
				Name:    NewRawIdent("PRIMARY KEY"),
				Columns: []*ColumnIdent{{Ident: column.Name}},
			})
		case TOKEN_REFERENCES:
			if err := p.checkPeekToken(TOKEN_IDENT); err != nil {
				return nil, apperr.Errorf("checkPeekToken: %w", err)
			}
			p.nextToken() // current = table_name
			constraint := &ForeignKeyConstraint{
				Name:    NewRawIdent(fmt.Sprintf("%s_%s_fkey", tableName.StringForDiff(), column.Name.StringForDiff())),
				Ref:     NewRawIdent(p.currentToken.Literal.Str),
				Columns: []*ColumnIdent{{Ident: column.Name}},
			}
			p.nextToken() // current = (
			idents, err := p.parseColumnIdents()
			if err != nil {
				return nil, apperr.Errorf("parseColumnIdents: %w", err)
			}
			constraint.RefColumns = idents
			constraints = constraints.Append(constraint)
		case TOKEN_UNIQUE:
			constraints = constraints.Append(&IndexConstraint{
				Unique:  true,
				Name:    NewRawIdent(fmt.Sprintf("%s_unique_%s", tableName.StringForDiff(), column.Name.StringForDiff())),
				Columns: []*ColumnIdent{{Ident: column.Name}},
			})
		case TOKEN_CHECK:
			if err := p.checkPeekToken(TOKEN_OPEN_PAREN); err != nil {
				return nil, apperr.Errorf("checkPeekToken: %w", err)
			}
			p.nextToken() // current = (
			constraint := &CheckConstraint{
				Name: NewRawIdent(fmt.Sprintf("%s_%s_check", tableName.StringForDiff(), column.Name.StringForDiff())),
			}
			idents, err := p.parseExpr()
			if err != nil {
				return nil, apperr.Errorf("parseExpr: %w", err)
			}
			constraint.Expr = constraint.Expr.Append(idents...)
			constraints = constraints.Append(constraint)
		case TOKEN_IDENT, TOKEN_COMMA, TOKEN_CLOSE_PAREN:
			break LabelConstraints
		default:
			return nil, apperr.Errorf("currentToken=%#v: %w", p.currentToken, ddl.ErrUnexpectedToken)
		}

		p.nextToken()
	}

	return constraints, nil
}

//nolint:funlen,cyclop,gocognit
func (p *Parser) parseTableConstraint(tableName *Ident) (Constraint, error) { //nolint:ireturn
	var constraintName *Ident
	if p.isCurrentToken(TOKEN_CONSTRAINT) {
		p.nextToken() // current = constraint_name
		if p.currentToken.Type != TOKEN_IDENT {
			return nil, apperr.Errorf("currentToken=%#v: %w", p.currentToken, ddl.ErrUnexpectedToken)
		}
		constraintName = NewRawIdent(p.currentToken.Literal.Str)
		p.nextToken() // current = PRIMARY or CHECK
	}

	switch p.currentToken.Type { //nolint:exhaustive
	case TOKEN_PRIMARY:
		if err := p.checkPeekToken(TOKEN_KEY); err != nil {
			return nil, apperr.Errorf("checkPeekToken: %w", err)
		}
		p.nextToken() // current = KEY
		if err := p.checkPeekToken(TOKEN_OPEN_PAREN); err != nil {
			return nil, apperr.Errorf("checkPeekToken: %w", err)
		}
		p.nextToken() // current = (
		idents, err := p.parseColumnIdents()
		if err != nil {
			return nil, apperr.Errorf("parseColumnIdents: %w", err)
		}
		return &PrimaryKeyConstraint{
			Name:    NewRawIdent("PRIMARY KEY"),
			Columns: idents,
		}, nil
	case TOKEN_FOREIGN:
		if err := p.checkPeekToken(TOKEN_KEY); err != nil {
			return nil, apperr.Errorf("checkPeekToken: %w", err)
		}
		p.nextToken() // current = KEY
		if err := p.checkPeekToken(TOKEN_OPEN_PAREN); err != nil {
			return nil, apperr.Errorf("checkPeekToken: %w", err)
		}
		p.nextToken() // current = (
		idents, err := p.parseColumnIdents()
		if err != nil {
			return nil, apperr.Errorf("parseColumnIdents: %w", err)
		}
		if err := p.checkCurrentToken(TOKEN_REFERENCES); err != nil {
			return nil, apperr.Errorf("checkPeekToken: %w", err)
		}
		p.nextToken() // current = ref_table_name
		if err := p.checkCurrentToken(TOKEN_IDENT); err != nil {
			return nil, apperr.Errorf("checkCurrentToken: %w", err)
		}
		refName := NewRawIdent(p.currentToken.Literal.Str)

		p.nextToken() // current = (
		identsRef, err := p.parseColumnIdents()
		if err != nil {
			return nil, apperr.Errorf("parseColumnIdents: %w", err)
		}
		if constraintName == nil {
			name := tableName.StringForDiff()
			for _, ident := range idents {
				name += fmt.Sprintf("_%s", ident.StringForDiff())
			}
			name += "_fkey"
			constraintName = NewRawIdent(name)
		}
		return &ForeignKeyConstraint{
			Name:       constraintName,
			Columns:    idents,
			Ref:        refName,
			RefColumns: identsRef,
		}, nil

	case TOKEN_UNIQUE, TOKEN_INDEX, TOKEN_KEY:
		c := &IndexConstraint{}
		if p.isCurrentToken(TOKEN_UNIQUE) {
			c.Unique = true
			p.nextToken() // current = KEY or INDEX
		}
		if err := p.checkCurrentToken(TOKEN_INDEX, TOKEN_KEY); err != nil {
			return nil, apperr.Errorf("checkCurrentToken: %w", err)
		}
		p.nextToken() // current = index_name
		if err := p.checkCurrentToken(TOKEN_IDENT); err != nil {
			return nil, apperr.Errorf("checkCurrentToken: %w", err)
		}
		constraintName := NewRawIdent(p.currentToken.Literal.Str)
		if err := p.checkPeekToken(TOKEN_OPEN_PAREN); err != nil {
			return nil, apperr.Errorf("checkPeekToken: %w", err)
		}
		p.nextToken() // current = (
		idents, err := p.parseColumnIdents()
		if err != nil {
			return nil, apperr.Errorf("parseColumnIdents: %w", err)
		}
		c.Name = constraintName
		c.Columns = idents
		return c, nil
	case TOKEN_CHECK:
		constraint := &CheckConstraint{}
		if err := p.checkPeekToken(TOKEN_OPEN_PAREN); err != nil {
			return nil, apperr.Errorf("checkPeekToken: %w", err)
		}
		p.nextToken() // current = (
		idents, err := p.parseExpr()
		if err != nil {
			return nil, apperr.Errorf("parseExpr: %w", err)
		}
		if constraintName == nil {
			// TODO: handle CONSTRAINT name
			constraintName = NewRawIdent(tableName.StringForDiff() + "_chk")
		}
		constraint.Name = constraintName
		constraint.Expr = constraint.Expr.Append(idents...)
		return constraint, nil
	default:
		return nil, apperr.Errorf("currentToken=%s: %w", p.currentToken.Type, ddl.ErrUnexpectedToken)
	}
}

//nolint:cyclop,funlen
func (p *Parser) parseDataType() (*DataType, error) {
	dataType := &DataType{Type: TOKEN_ILLEGAL}

	switch p.currentToken.Type { //nolint:exhaustive
	case TOKEN_BOOLEAN:
		dataType.Name = "TINYINT"
		dataType.Type = TOKEN_BOOLEAN
		dataType.Expr = dataType.Expr.Append(NewRawIdent("1"))
	case TOKEN_TIMESTAMP:
		dataType.Name = p.currentToken.Literal.String()
		dataType.Type = TOKEN_TIMESTAMP
	case TOKEN_DATE:
		dataType.Name = p.currentToken.Literal.String()
		dataType.Type = TOKEN_DATE
	case TOKEN_TIME:
		dataType.Name = p.currentToken.Literal.String()
		dataType.Type = TOKEN_TIME
	case TOKEN_DATETIME:
		dataType.Name = p.currentToken.Literal.String()
		dataType.Type = TOKEN_DATETIME
	case TOKEN_DOUBLE:
		dataType.Name = p.currentToken.Literal.String()
		if err := p.checkPeekToken(TOKEN_PRECISION); err != nil {
			return nil, apperr.Errorf("checkPeekToken: %w", err)
		}
		p.nextToken() // current = PRECISION
		dataType.Name += " " + p.currentToken.Literal.String()
		dataType.Type = TOKEN_DOUBLE_PRECISION
	case TOKEN_CHARACTER:
		dataType.Name = p.currentToken.Literal.String()
		if err := p.checkPeekToken(TOKEN_VARYING); err != nil {
			return nil, apperr.Errorf("checkPeekToken: %w", err)
		}
		p.nextToken() // current = VARYING
		dataType.Name += " " + p.currentToken.Literal.String()
		dataType.Type = TOKEN_VARCHAR
	case TOKEN_ENUM:
		dataType.Name = p.currentToken.Literal.String()
		dataType.Type = TOKEN_ENUM
	default:
		dataType.Name = p.currentToken.Literal.String()
		dataType.Type = p.currentToken.Type
	}

	if p.isPeekToken(TOKEN_OPEN_PAREN) {
		p.nextToken() // current = (
		idents, err := p.parseIdents()
		if err != nil {
			return nil, apperr.Errorf("parseIdents: %w", err)
		}
		dataType.Expr = dataType.Expr.Append(idents...)
	}

	return dataType, nil
}

func (p *Parser) parseColumnIdents() ([]*ColumnIdent, error) {
	idents := make([]*ColumnIdent, 0)

LabelIdents:
	for {
		switch p.currentToken.Type { //nolint:exhaustive
		case TOKEN_OPEN_PAREN:
			// do nothing
		case TOKEN_IDENT:
			ident := &ColumnIdent{Ident: NewRawIdent(p.currentToken.Literal.Str)}
			switch p.peekToken.Type { //nolint:exhaustive
			case TOKEN_ASC:
				ident.Order = &Order{Desc: false}
				p.nextToken() // current = ASC
			case TOKEN_DESC:
				ident.Order = &Order{Desc: true}
				p.nextToken() // current = DESC
			default:
				// do nothing
			}
			idents = append(idents, ident)
		case TOKEN_COMMA:
			// do nothing
		case TOKEN_CLOSE_PAREN:
			p.nextToken()
			break LabelIdents
		default:
			return nil, apperr.Errorf("currentToken=%#v: %w", p.currentToken, ddl.ErrUnexpectedToken)
		}
		p.nextToken()
	}

	return idents, nil
}

func (p *Parser) parseIdents() ([]*Ident, error) {
	idents := make([]*Ident, 0)

LabelIdents:
	for {
		switch p.currentToken.Type { //nolint:exhaustive
		case TOKEN_OPEN_PAREN:
			// do nothing
		case TOKEN_IDENT:
			idents = append(idents, NewRawIdent(p.currentToken.Literal.Str))
		case TOKEN_CLOSE_PAREN:
			break LabelIdents
		case TOKEN_EOF, TOKEN_ILLEGAL:
			return nil, apperr.Errorf("currentToken=%#v: %w", p.currentToken, ddl.ErrUnexpectedToken)
		default:
			idents = append(idents, NewRawIdent(p.currentToken.Literal.Str))
		}
		p.nextToken()
	}

	return idents, nil
}

// MEMO: backup
// TODO: check if this is necessary
// func isOperator(tokenType TokenType) bool {
// 	switch tokenType { //nolint:exhaustive
// 	case TOKEN_EQUAL, TOKEN_GREATER, TOKEN_LESS,
// 		TOKEN_PLUS, TOKEN_MINUS, TOKEN_ASTERISK, TOKEN_SLASH,
// 		TOKEN_TYPE_ANNOTATION,
// 		TOKEN_STRING_CONCAT, TOKEN_TYPECAST:
// 		return true
// 	default:
// 		return false
// 	}
// }

func isReservedValue(tokenType TokenType) bool {
	switch tokenType { //nolint:exhaustive
	case TOKEN_NULL, TOKEN_TRUE, TOKEN_FALSE:
		return true
	default:
		return false
	}
}

func isDataType(tokenType TokenType) bool {
	switch tokenType { //nolint:exhaustive
	case TOKEN_BOOLEAN,
		TOKEN_TINYINT,
		TOKEN_SMALLINT, TOKEN_INTEGER, TOKEN_BIGINT,
		TOKEN_DECIMAL, TOKEN_NUMERIC,
		TOKEN_REAL, TOKEN_DOUBLE, /* TOKEN_PRECISION, */
		TOKEN_SMALLSERIAL, TOKEN_SERIAL, TOKEN_BIGSERIAL,
		TOKEN_JSON,
		TOKEN_CHARACTER, TOKEN_VARYING,
		TOKEN_VARCHAR, TOKEN_TEXT,
		TOKEN_TIMESTAMP, TOKEN_DATE, TOKEN_TIME,
		TOKEN_DATETIME,
		TOKEN_ENUM:
		return true
	default:
		return false
	}
}

func isConstraint(tokenType TokenType) bool {
	switch tokenType { //nolint:exhaustive
	case TOKEN_CONSTRAINT,
		TOKEN_INDEX,
		TOKEN_PRIMARY, TOKEN_KEY,
		TOKEN_FOREIGN, TOKEN_REFERENCES,
		TOKEN_UNIQUE,
		TOKEN_CHECK:
		return true
	default:
		return false
	}
}

func (p *Parser) isCurrentToken(expectedTypes ...TokenType) bool {
	for _, expected := range expectedTypes {
		if expected == p.currentToken.Type {
			return true
		}
	}
	return false
}

func (p *Parser) checkCurrentToken(expectedTypes ...TokenType) error {
	for _, expected := range expectedTypes {
		if expected == p.currentToken.Type {
			return nil
		}
	}
	return apperr.Errorf("currentToken: expected=%s, but got=%#v: %w", stringz.JoinStringers(",", expectedTypes...), p.currentToken, ddl.ErrUnexpectedToken)
}

func (p *Parser) isPeekToken(expectedTypes ...TokenType) bool {
	for _, expected := range expectedTypes {
		if expected == p.peekToken.Type {
			return true
		}
	}
	return false
}

func (p *Parser) checkPeekToken(expectedTypes ...TokenType) error {
	for _, expected := range expectedTypes {
		if expected == p.peekToken.Type {
			return nil
		}
	}
	return apperr.Errorf("peekToken: expected=%s, but got=%#v: %w", stringz.JoinStringers(",", expectedTypes...), p.peekToken, ddl.ErrUnexpectedToken)
}
