package godac

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

// ColSepWide is column names separator.
const ColSepWide = ", "

// Placeholder is sql query placeholder.
const Placeholder = "?"

// Table is a sql database table.
type Table struct {
	Name   string
	Fields []Field
}

// Select query sql SELECT.
func (table *Table) Select(db DB, clauses string, args ...interface{}) ([]Map, error) {
	if table.Name == "" {
		return nil, errors.New("Table name cannot be empty")
	}

	var cols []string
	for _, field := range table.Fields {
		cols = append(cols, field.Name)
	}
	if len(cols) == 0 {
		cols = append(cols, "*")
	}

	query := fmt.Sprintf("SELECT %s from %s %s", strings.Join(cols, ColSepWide), table.Name, clauses)
	return MapQuery(db, query, args...)
}

// Insert execute sql INSERT INTO.
func (table *Table) Insert(db DB, record Map) (sql.Result, error) {
	var value interface{}
	var cols []string
	var placeholders []string
	var args []interface{}

	for _, field := range table.Fields {
		if field.IsAutoInc {
			continue
		}
		if field.ReadOnly {
			if field.Default == nil {
				continue
			} else {
				value = field.Default
			}
		} else {
			value = record[field.Name]
			if value == nil {
				value = field.Default
			}
			if value == nil {
				continue
			}
		}
		for _, rule := range field.Validations {
			if err := rule.Validate(value); err != nil {
				return nil, err
			}
		}
		cols = append(cols, field.Name)
		placeholders = append(placeholders, Placeholder)
		args = append(args, value)
	}

	query := "INSERT INTO %s(%s)VALUES(%s)"
	query = fmt.Sprintf(query, table.Name, strings.Join(cols, ColSepWide), strings.Join(placeholders, ColSepWide))
	return db.Exec(query, args...)
}

// Update execute sql UPDATE.
func (table *Table) Update(db DB, record Map) (sql.Result, error) {
	var sets []string
	var args []interface{}

	for _, field := range table.Fields {
		if field.InPrimaryKey || field.IsAutoInc {
			continue
		}
		value, exist := record[field.Name]
		if !(!field.ReadOnly && exist) {
			if field.OnUpdate == nil {
				continue
			} else {
				value = field.OnUpdate
			}
		}
		for _, rule := range field.Validations {
			if err := rule.Validate(value); err != nil {
				return nil, err
			}
		}
		sets = append(sets, fmt.Sprintf("%s = %s", field.Name, Placeholder))
		args = append(args, value)
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
	var condition []string
	for _, field := range table.Fields {
		if !field.InPrimaryKey {
			continue
		}
		value, exist := record[field.Name]
		if !exist {
			err = fmt.Errorf("Key %s not exists", field.Name)
			return
		}
		condition = append(condition, fmt.Sprintf("%s = %s", field.Name, Placeholder))
		args = append(args, value)
	}
	query = strings.Join(condition, " AND ")
	return
}

// Count query SELECT COUNT(*).
func (table *Table) Count(db DB, where string, args ...interface{}) (int64, error) {
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s", table.Name)
	if where != "" {
		query += " WHERE " + where
	}

	row := db.QueryRow(query, args...)
	var count int64
	if err := row.Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

// CountValue query SELECT COUNT(*) by column value. used for detect duplicate value.
func (table *Table) CountValue(db DB, column string, value interface{}, where string, args ...interface{}) (int64, error) {
	condition := "= ?"
	if value == nil {
		condition = "IS NULL"
	} else {
		switch value.(type) {
		case string:
			value = strings.TrimSpace(value.(string))
		default:
		}
		args = append([]interface{}{value}, args...)
	}
	condition = column + " " + condition
	if where == "" {
		where = condition
	} else {
		where += condition + " AND " + where
	}

	return table.Count(db, where, args...)
}

// CountValueByRecord query SELECT COUNT(*) by record.
func (table *Table) CountValueByRecord(db DB, column string, record Map, excludeThisRecord bool, where string, args ...interface{}) (int64, error) {
	if excludeThisRecord {
		queryPK, argsPK, err := table.WherePrimaryKey(record)
		if err != nil {
			return 0, err
		}
		queryPK = strings.Replace(queryPK, "=", "<>", 0)
		if where == "" {
			where = queryPK
		} else {
			where = queryPK + " AND " + where
		}
		args = append(argsPK, args...)
	}
	return table.CountValue(db, column, record[column], where, args...)
}
