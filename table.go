package godac

import (
	"errors"
	"fmt"
	"godac/sqlbuilder"
	"strings"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

// Placeholder is sql query placeholder.
const Placeholder = "?"

// Table is a sql database table.
type Table struct {
	active     bool
	cols       []string
	keys       []string
	keysMap    map[string]string
	primaryKey []int // Indexes of primary key fields.
	autoInc    int   // index of AutoInc field.

	Name     string
	Fields   []Field
	OnInsert ActionFunc
	OnUpdate ActionFunc
	OnDelete ActionFunc
}

// Open init the Table.
func (table *Table) Open() error {
	if table.active {
		return nil
	}
	if strings.TrimSpace(table.Name) == "" {
		return errors.New("Table name cannot be empty")
	}
	table.cols = []string{}
	table.keys = []string{}
	table.keysMap = map[string]string{}
	table.primaryKey = []int{}
	table.autoInc = -1
	for i, field := range table.Fields {
		if strings.TrimSpace(field.Name) == "" {
			return fmt.Errorf("Fields[%d]: name cannot be empty", i)
		}
		table.cols = append(table.cols, field.Name)
		key := field.GetKey()
		table.keys = append(table.keys, key)
		table.keysMap[field.Name] = key
		if field.PrimaryKey {
			table.primaryKey = append(table.primaryKey, i)
		}
		if table.autoInc < 0 && field.AutoInc {
			table.autoInc = i
		}
	}
	table.active = true
	return nil
}

// Close the table.
func (table *Table) Close() {
	table.active = false
}

// Select query sql SELECT.
func (table *Table) Select(db DB, selector sqlbuilder.Selector, args ...interface{}) ([]Map, error) {
	if err := table.Open(); err != nil {
		return nil, err
	}
	query := selector.Columns(table.cols...).From(table.Name).SQL()
	return MapQuery(table.keysMap, db, query, args...)
}

func (table *Table) execAction(c Context, onAction, defaultAction ActionFunc) (Result, error) {
	if err := table.Open(); err != nil {
		return nil, err
	}
	if onAction == nil {
		return defaultAction(c)
	}
	return onAction(c)
}

// Insert execute sql INSERT INTO.
func (table *Table) Insert(db DB, record Map) (Result, error) {
	c := Context{StateInsert, db, table, record, Field{}}
	return table.execAction(c, table.OnInsert, table.DefaultInsert)
}

// DefaultInsert is default Insert handler.
func (table *Table) DefaultInsert(c Context) (Result, error) {
	if err := table.Open(); err != nil {
		return nil, err
	}
	var rec = Map{}
	for k, v := range c.Record {
		rec[k] = v
	}
	var cols []string
	var placeholders []string
	var args []interface{}
	for i, field := range table.Fields {
		if field.AutoInc {
			continue
		}
		value := c.Record[table.keys[i]]
		if field.ReadOnly {
			if field.Default == nil {
				continue
			}
			value = nil
		}
		if value == nil {
			value = field.GetDefault()
		}
		rec[table.keys[i]] = value
		if err := validateField(Context{StateInsert, c.DB, table, rec, field}, value); err != nil {
			return nil, err
		}
		cols = append(cols, field.Name)
		placeholders = append(placeholders, Placeholder)
		args = append(args, value)
	}
	query := "INSERT INTO %s(%s)VALUES(%s)"
	query = fmt.Sprintf(query, table.Name, strings.Join(cols, sqlbuilder.ColSep), strings.Join(placeholders, sqlbuilder.ColSep))
	rst, err := c.DB.Exec(query, args...)
	return NewResult(Context{StateInsert, c.DB, table, rec, Field{}}, rst), err
}

// Update execute sql UPDATE.
func (table *Table) Update(db DB, record Map) (Result, error) {
	c := Context{StateUpdate, db, table, record, Field{}}
	return table.execAction(c, table.OnUpdate, table.DefaultUpdate)
}

// DefaultUpdate is default Update handler.
func (table *Table) DefaultUpdate(c Context) (Result, error) {
	whereQuery, whereArgs, err := table.WherePrimaryKey(false, false, c.Record)
	if err != nil {
		return nil, err
	}
	var rec = Map{}
	for k, v := range c.Record {
		rec[k] = v
	}
	var sets []string
	var args []interface{}
	for i, field := range table.Fields {
		if field.PrimaryKey || field.AutoInc {
			continue
		}
		value, exist := c.Record[table.keys[i]]
		if field.ReadOnly || !exist {
			if field.OnUpdate == nil {
				continue
			}
			value = nil
		}
		if value == nil {
			value = field.GetOnUpdate()
		}
		rec[table.keys[i]] = value
		if err := validateField(Context{StateUpdate, c.DB, table, rec, field}, value); err != nil {
			return nil, err
		}
		sets = append(sets, fmt.Sprintf("%s = %s", field.Name, Placeholder))
		args = append(args, value)
	}
	if len(sets) == 0 {
		return nil, fmt.Errorf("Table %s: not enough columns to update", table.Name)
	}
	query := "UPDATE %s SET %s WHERE %s"
	query = fmt.Sprintf(query, table.Name, strings.Join(sets, sqlbuilder.ColSep), whereQuery)
	rst, err := c.DB.Exec(query, append(args, whereArgs...)...)
	return NewResult(Context{StateUpdate, c.DB, table, rec, Field{}}, rst), err
}

// Validate field rules.
func validateField(c Context, value interface{}) error {
	field := c.Field
	for _, v := range field.Validations {
		if rule, ok := v.(ValidationRule); ok {
			rule.SetContext(c)
		}
	}
	if err := validation.Validate(value, field.Validations...); err != nil {
		if e, ok := err.(validation.Error); ok {
			return e.SetMessage(fmt.Sprintf(ErrorFormatOnValidation, field.GetTitle(), e))
		}
		return err
	}
	return nil
}

// Delete execute sql DELETE;
func (table *Table) Delete(db DB, record Map) (Result, error) {
	c := Context{StateDelete, db, table, record, Field{}}
	return table.execAction(c, table.OnDelete, table.DefaultDelete)
}

// DefaultDelete is default Delete handler.
func (table *Table) DefaultDelete(c Context) (Result, error) {
	query, args, err := table.WherePrimaryKey(false, false, c.Record)
	if err != nil {
		return nil, err
	}

	query = fmt.Sprintf("DELETE FROM %s WHERE %s", table.Name, query)
	rst, err := c.DB.Exec(query, args...)
	return NewResult(Context{StateDelete, c.DB, table, c.Record, Field{}}, rst), err
}

// WherePrimaryKey get where sql by primary key in record.
func (table *Table) WherePrimaryKey(hasTableName, reversed bool, record Map) (condition string, args []interface{}, err error) {
	if err = table.Open(); err != nil {
		return
	}
	if len(table.primaryKey) == 0 {
		err = fmt.Errorf("The table %s does not define primary key", table.Name)
		return
	}
	operator := "="
	if reversed {
		operator = "<>"
	}
	var conditions []string
	for _, i := range table.primaryKey {
		key := table.keys[i]
		value, exist := record[key]
		if !exist {
			err = fmt.Errorf("Primary key %s is required in record", key)
			return
		}
		fieldName := table.Fields[i].Name
		if hasTableName {
			fieldName = table.Name + "." + fieldName
		}
		conditions = append(conditions, fmt.Sprintf("%s %s %s", fieldName, operator, Placeholder))
		args = append(args, value)
	}
	condition = strings.Join(conditions, " AND ")
	return
}

// Count query SELECT COUNT(*).
func (table *Table) Count(db DB, selector sqlbuilder.Selector, args ...interface{}) (int64, error) {
	if err := table.Open(); err != nil {
		return 0, err
	}
	query := selector.Columns("COUNT(*)").From(table.Name).SQL()
	row := db.QueryRow(query, args...)
	var count int64
	if err := row.Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

// CountValue query SELECT COUNT(*) by column value. used for detect duplicate value.
func (table *Table) CountValue(db DB, field Field, value interface{}, selector sqlbuilder.Selector, args ...interface{}) (int64, error) {
	column := field.Name
	condition := "= ?"
	if value == nil {
		condition = "IS NULL"
	} else {
		switch v := value.(type) {
		case string:
			column = fmt.Sprintf("TRIM(%s)", column)
			value = strings.TrimSpace(v)
		default:
		}
		args = append(args, value)
	}
	condition = column + " " + condition
	selector = selector.WhereAnd(condition)
	return table.Count(db, selector, args...)
}

// CountRecord query SELECT COUNT(*) by record.
func (table *Table) CountRecord(db DB, field Field, record Map, excludeSelf bool, selector sqlbuilder.Selector, args ...interface{}) (int64, error) {
	if excludeSelf {
		queryPK, argsPK, err := table.WherePrimaryKey(false, true, record)
		if err != nil {
			return 0, err
		}
		selector = selector.WhereAnd(queryPK)
		args = append(args, argsPK...)
	}
	if err := table.Open(); err != nil {
		return 0, err
	}
	return table.CountValue(db, field, record[table.keysMap[field.Name]], selector, args...)
}
