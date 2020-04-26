package sqlbuilder

import (
	"fmt"
	"strings"
)

// ColSep is column names separator.
const ColSep = ", "

// Select create a SELECT SQL builder.
func Select() Selector {
	return Selector{}
}

// Selector is a sql builder for SELECT.
type Selector struct {
	columns, joins       []string
	from, where, orderBy string
	limit, offset        *int64
}

// Merge source to target
func (sql Selector) Merge(src Selector) Selector {
	if len(src.columns) > 0 {
		sql.columns = src.columns
	}
	if len(src.joins) > 0 {
		sql.joins = src.joins
	}
	if src.from != "" {
		sql.from = src.from
	}
	if src.where != "" {
		sql.where = src.where
	}
	if src.orderBy != "" {
		sql.orderBy = src.orderBy
	}
	if src.limit != nil {
		sql.limit = src.limit
	}
	if src.offset != nil {
		sql.offset = src.offset
	}
	return sql
}

// Columns set columns.
func (sql Selector) Columns(columns ...string) Selector {
	sql.columns = columns
	return sql
}

// From set FROM clause.
func (sql Selector) From(from string) Selector {
	sql.from = from
	return sql
}

// LeftJoin set LEFT JOIN clause.
func (sql Selector) LeftJoin(joined, on string) Selector {
	sql.joins = append(sql.joins, fmt.Sprintf("LEFT JOIN %s ON %s", joined, on))
	return sql
}

// Where set WHERE clause.
func (sql Selector) Where(where string) Selector {
	sql.where = where
	return sql
}

// WhereAnd set WHERE + AND clause.
func (sql Selector) WhereAnd(where string) Selector {
	if sql.where == "" {
		return sql.Where(where)
	}
	sql.where += " AND " + where
	return sql
}

// OrderBy set ORDER BY clause.
func (sql Selector) OrderBy(orderBy string) Selector {
	sql.orderBy = orderBy
	return sql
}

// Limit set LIMIT clause.
func (sql Selector) Limit(limit int64) Selector {
	sql.limit = &limit
	return sql
}

// Offset set OFFSET clause.
func (sql Selector) Offset(offset int64) Selector {
	sql.offset = &offset
	return sql
}

func iifAdd(s string, expr bool, t, f string) string {
	if expr {
		return s + t
	}
	return s + f
}

// SQL get real sql.
func (sql Selector) SQL() string {
	s := "SELECT "
	s = iifAdd(s, len(sql.columns) == 0, "*", strings.Join(sql.columns, ColSep))
	s = iifAdd(s, sql.from != "", " FROM "+sql.from, "")
	if sql.from != "" && len(sql.joins) > 0 {
		s += " " + strings.Join(sql.joins, " ")
	}
	s = iifAdd(s, sql.where != "", " WHERE "+sql.where, "")
	s = iifAdd(s, sql.orderBy != "", " ORDER BY "+sql.orderBy, "")
	if sql.limit != nil {
		s += fmt.Sprintf(" LIMIT %d", *sql.limit)
	}
	if sql.offset != nil {
		s += fmt.Sprintf(" OFFSET %d", *sql.offset)
	}
	return s
}
