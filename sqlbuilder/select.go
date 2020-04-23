package sqlbuilder

import "strings"

// ColSep is column names separator.
const ColSep = ", "

// Selector is a sql builder for SELECT.
type Selector struct {
	columns              []string
	from, where, orderBy string
}

// Select create a SELECT SQL builder.
func Select() Selector {
	return Selector{}
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
	s = iifAdd(s, sql.where != "", " WHERE "+sql.where, "")
	s = iifAdd(s, sql.orderBy != "", " ORDER BY "+sql.orderBy, "")
	return s
}
