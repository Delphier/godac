package godac

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

// ColSepWide is column names separator.
const ColSepWide = ", "

// Placeholder is sql query placeholder.
const Placeholder = "?"

// Table is a sql database table.
type Table struct {
	active     bool
	cols       string
	keys       []string
	keysMap    map[string]string
	primaryKey []int // Indexes of primary key fields.
	autoInc    int   // index of AutoInc field.

	Name     string
	Fields   []Field
	OnInsert func(*Table, DB, Map) (Result, error)
	OnUpdate func(*Table, DB, Map) (Result, error)
	OnDelete func(*Table, DB, Map) (sql.Result, error)
}

// Open init the Table.
func (table *Table) Open() error {
	if table.active {
		return nil
	}
	if strings.TrimSpace(table.Name) == "" {
		return errors.New("Table name cannot be empty")
	}
	table.cols = ""
	table.keys = []string{}
	table.keysMap = map[string]string{}
	table.primaryKey = []int{}
	table.autoInc = -1
	for i, field := range table.Fields {
		if strings.TrimSpace(field.Name) == "" {
			return fmt.Errorf("Fields[%d]: name cannot be empty", i)
		}
		if table.cols != "" {
			table.cols = table.cols + ColSepWide
		}
		table.cols = table.cols + field.Name
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
	if table.cols == "" {
		table.cols = "*"
	}
	table.active = true
	return nil
}

// Close the table.
func (table *Table) Close() {
	table.active = false
}

// Select query sql SELECT.
func (table *Table) Select(db DB, clauses string, args ...interface{}) ([]Map, error) {
	if err := table.Open(); err != nil {
		return nil, err
	}
	query := fmt.Sprintf("SELECT %s FROM %s %s", table.cols, table.Name, clauses)
	return MapQuery(table.keysMap, db, query, args...)
}

// Insert execute sql INSERT INTO.
func (table *Table) Insert(db DB, record Map) (Result, error) {
	if err := table.Open(); err != nil {
		return nil, err
	}
	if table.OnInsert == nil {
		return table.DefaultInsert(db, record)
	}
	return table.OnInsert(table, db, record)
}

// DefaultInsert is default Insert handler.
func (table *Table) DefaultInsert(db DB, record Map) (Result, error) {
	if err := table.Open(); err != nil {
		return nil, err
	}
	var rec = Map{}
	for k, v := range record {
		rec[k] = v
	}
	var cols []string
	var placeholders []string
	var args []interface{}
	for i, field := range table.Fields {
		if field.AutoInc {
			continue
		}
		value := record[table.keys[i]]
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
		if err := table.validateField(db, field, rec, value, true); err != nil {
			return nil, err
		}
		cols = append(cols, field.Name)
		placeholders = append(placeholders, Placeholder)
		args = append(args, value)
	}
	query := "INSERT INTO %s(%s)VALUES(%s)"
	query = fmt.Sprintf(query, table.Name, strings.Join(cols, ColSepWide), strings.Join(placeholders, ColSepWide))
	rst, err := db.Exec(query, args...)
	return result{rst, true, db, table, rec}, err
}

// Update execute sql UPDATE.
func (table *Table) Update(db DB, record Map) (Result, error) {
	if err := table.Open(); err != nil {
		return nil, err
	}
	if table.OnUpdate == nil {
		return table.DefaultUpdate(db, record)
	}
	return table.OnUpdate(table, db, record)
}

// DefaultUpdate is default Update handler.
func (table *Table) DefaultUpdate(db DB, record Map) (Result, error) {
	whereQuery, whereArgs, err := table.WherePrimaryKey(record)
	if err != nil {
		return nil, err
	}
	var rec = Map{}
	for k, v := range record {
		rec[k] = v
	}
	var sets []string
	var args []interface{}
	for i, field := range table.Fields {
		if field.PrimaryKey || field.AutoInc {
			continue
		}
		value, exist := record[table.keys[i]]
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
		if err := table.validateField(db, field, rec, value, false); err != nil {
			return nil, err
		}
		sets = append(sets, fmt.Sprintf("%s = %s", field.Name, Placeholder))
		args = append(args, value)
	}
	if len(sets) == 0 {
		return nil, fmt.Errorf("Table %s: not enough columns to update", table.Name)
	}
	query := "UPDATE %s SET %s WHERE %s"
	query = fmt.Sprintf(query, table.Name, strings.Join(sets, ColSepWide), whereQuery)
	rst, err := db.Exec(query, append(args, whereArgs...)...)
	return result{rst, false, db, table, rec}, err
}

// ValidationErrorFormat define field validation error format.
// %s represents field title, %v represents validation error, field must be in front of the error.
var ValidationErrorFormat = "%s: %v"

// Validate field rules.
func (table *Table) validateField(db DB, field Field, record Map, value interface{}, isInsert bool) error {
	for _, v := range field.Validations {
		if rule, ok := v.(ValidationRule); ok {
			rule.Init(db, table, field, record, isInsert)
		}
	}
	if err := validation.Validate(value, field.Validations...); err != nil {
		if e, ok := err.(validation.Error); ok {
			return e.SetMessage(fmt.Sprintf(ValidationErrorFormat, field.GetTitle(), e))
		}
		return err
	}
	return nil
}

// Delete execute sql DELETE;
func (table *Table) Delete(db DB, record Map) (sql.Result, error) {
	if err := table.Open(); err != nil {
		return nil, err
	}
	if table.OnDelete == nil {
		return table.DefaultDelete(db, record)
	}
	return table.OnDelete(table, db, record)
}

// DefaultDelete is default Delete handler.
func (table *Table) DefaultDelete(db DB, record Map) (sql.Result, error) {
	query, args, err := table.WherePrimaryKey(record)
	if err != nil {
		return nil, err
	}

	query = fmt.Sprintf("DELETE FROM %s WHERE %s", table.Name, query)
	return db.Exec(query, args...)
}

// WherePrimaryKey get where sql by primary key in record.
func (table *Table) WherePrimaryKey(record Map) (query string, args []interface{}, err error) {
	if err = table.Open(); err != nil {
		return
	}
	if len(table.primaryKey) == 0 {
		err = fmt.Errorf("The table %s does not define primary key", table.Name)
		return
	}
	var condition []string
	for _, i := range table.primaryKey {
		key := table.keys[i]
		value, exist := record[key]
		if !exist {
			err = fmt.Errorf("Primary key %s is required in record", key)
			return
		}
		condition = append(condition, fmt.Sprintf("%s = %s", table.Fields[i].Name, Placeholder))
		args = append(args, value)
	}
	query = strings.Join(condition, " AND ")
	return
}

// Count query SELECT COUNT(*).
func (table *Table) Count(db DB, where string, args ...interface{}) (int64, error) {
	if err := table.Open(); err != nil {
		return 0, err
	}
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
func (table *Table) CountValue(db DB, field Field, value interface{}, where string, args ...interface{}) (int64, error) {
	column := field.Name
	condition := "= ?"
	if value == nil {
		condition = "IS NULL"
	} else {
		switch value.(type) {
		case string:
			column = fmt.Sprintf("TRIM(%s)", column)
			value = strings.TrimSpace(value.(string))
		default:
		}
		args = append([]interface{}{value}, args...)
	}
	condition = column + " " + condition
	if where == "" {
		where = condition
	} else {
		where = condition + " AND " + where
	}

	return table.Count(db, where, args...)
}

// CountRecord query SELECT COUNT(*) by record.
func (table *Table) CountRecord(db DB, field Field, record Map, excludeSelf bool, where string, args ...interface{}) (int64, error) {
	if excludeSelf {
		queryPK, argsPK, err := table.WherePrimaryKey(record)
		if err != nil {
			return 0, err
		}
		queryPK = strings.Replace(queryPK, "=", "<>", -1)
		if where == "" {
			where = queryPK
		} else {
			where = queryPK + " AND " + where
		}
		args = append(argsPK, args...)
	}
	if err := table.Open(); err != nil {
		return 0, err
	}
	return table.CountValue(db, field, record[table.keysMap[field.Name]], where, args...)
}
