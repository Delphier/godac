package sqlexpress

import (
	"database/sql"
	"fmt"
	"strings"
)

// ColSep and ColSepWide is column names separator.
const (
	ColSep     = ","
	ColSepWide = ColSep + " "
)

const placeholder = "?"

// TableProfile is a sql database table profile.
type TableProfile struct {
	Name                    string // Table name
	PrimaryKey              string // Multi-column supports
	AutoIncColumn           string
	ExcludedColumnsOnInsert string
	ExcludedColumnsOnUpdate string
	OtherColumns            string
}

// Table is a sql database table.
type Table struct {
	Name             string
	PrimaryKey       []string
	AutoIncColumn    string
	ColumnsForSelect []string
	ColumnsForInsert []string
	ColumnsForUpdate []string
}

// NewTable create Table instance from *TableProfile.
func NewTable(profile *TableProfile) *Table {
	result := Table{Name: profile.Name, AutoIncColumn: profile.AutoIncColumn}
	result.PrimaryKey = trimColumns(strings.Split(profile.PrimaryKey, ColSep))

	cols := profile.PrimaryKey + ColSep + profile.OtherColumns
	result.ColumnsForSelect = trimColumns(strings.Split(cols, ColSep))
	cols = profile.AutoIncColumn + ColSep + profile.ExcludedColumnsOnUpdate + ColSep + profile.ExcludedColumnsOnInsert
	result.ColumnsForSelect = addColumns(cols, result.ColumnsForSelect)

	cols = profile.AutoIncColumn + ColSep + profile.ExcludedColumnsOnInsert
	result.ColumnsForInsert = removeColumns(cols, result.ColumnsForSelect)

	cols = profile.PrimaryKey + ColSep + profile.AutoIncColumn + ColSep + profile.ExcludedColumnsOnUpdate
	result.ColumnsForUpdate = removeColumns(cols, result.ColumnsForSelect)

	return &result
}

// Select query sql SELECT.
func (table *Table) Select(db DB, clauses string, args ...interface{}) ([]Map, error) {
	query := strings.Join(table.ColumnsForSelect, ColSepWide)
	if query == "" {
		query = "*"
	}
	query = fmt.Sprintf("SELECT %s from %s %s", query, table.Name, clauses)
	return MapQuery(db, query, args...)
}

// Insert execute sql INSERT INTO.
func (table *Table) Insert(db DB, record Map) (sql.Result, error) {
	var cols []string
	var placeholders []string
	var args []interface{}

	for _, col := range table.ColumnsForInsert {
		value, exist := record[col]
		if exist {
			cols = append(cols, col)
			placeholders = append(placeholders, placeholder)
			args = append(args, value)
		}
	}

	query := "INSERT INTO %s(%s)VALUES(%s)"
	query = fmt.Sprintf(query, table.Name, strings.Join(cols, ColSepWide), strings.Join(placeholders, ColSepWide))
	return db.Exec(query, args...)
}

// Update execute sql UPDATE.
func (table *Table) Update(db DB, record Map) (sql.Result, error) {
	var sets []string
	var args []interface{}

	for _, col := range table.ColumnsForUpdate {
		value, exist := record[col]
		if exist {
			sets = append(sets, col+" = "+placeholder)
			args = append(args, value)
		}
	}

	whereQuery, whereArgs, err := table.WherePrimaryKey(record)
	if err != nil {
		return nil, err
	}

	query := "UPDATE %s SET %s WHERE %s"
	query = fmt.Sprintf(query, table.Name, strings.Join(sets, ColSepWide), whereQuery)
	return db.Exec(query, append(args, whereArgs...)...)
}

// Delete execute sql DELETE;
func (table *Table) Delete(db DB, record Map) (sql.Result, error) {
	query, args, err := table.WherePrimaryKey(record)
	if err != nil {
		return nil, err
	}

	query = fmt.Sprintf("DELETE FROM %s WHERE %s", table.Name, query)
	return db.Exec(query, args...)
}

// WherePrimaryKey get where sql by primary key in record.
func (table *Table) WherePrimaryKey(record Map) (query string, args []interface{}, err error) {
	keys := append([]string{}, table.PrimaryKey...)
	for _, key := range keys {
		value, exist := record[key]
		if !exist {
			err = fmt.Errorf("key %s not exists", key)
			return
		}
		args = append(args, value)
	}
	for i := range keys {
		keys[i] = fmt.Sprintf("%s = %s", keys[i], placeholder)
	}
	query = strings.Join(keys, " AND ")
	return
}

// trimColumns trim column name and remove empty column.
func trimColumns(ss []string) []string {
	for i := len(ss) - 1; i >= 0; i-- {
		ss[i] = strings.TrimSpace(ss[i])
		if ss[i] == "" {
			ss = append(ss[:i], ss[i+1:]...)
		}
	}
	return ss
}

// columnExists checks whether the column exists.
func columnExists(col string, cols []string) bool {
	for _, s := range cols {
		if s == col {
			return true
		}
	}
	return false
}

// addColumns add columns to src.
func addColumns(cols string, src []string) []string {
	result := append([]string{}, src...)
	for _, col := range strings.Split(cols, ColSep) {
		col = strings.TrimSpace(col)
		if col != "" && !columnExists(col, result) {
			result = append(result, col)
		}
	}
	return result
}

// removeColumns remove columns from src.
func removeColumns(cols string, src []string) []string {
	result := append([]string{}, src...)
	for _, col := range strings.Split(cols, ColSep) {
		col = strings.TrimSpace(col)
		for i := len(result) - 1; i >= 0; i-- {
			if result[i] == col {
				result = append(result[:i], result[i+1:]...)
			}
		}
	}
	return result
}
