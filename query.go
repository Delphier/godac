package godac

import (
	"errors"
	"godac/sqlbuilder"
)

// Query represents a table created by sql query.
type Query struct {
	active       bool
	defaultTable *Table
	keysMap      map[string]string

	Selector sqlbuilder.Selector
	Tables   []*Table
	Fields   []Field
	OnInsert ActionFunc
	OnUpdate ActionFunc
	OnDelete ActionFunc
}

// Open to init the query.
func (query *Query) Open() {
	if query.active {
		return
	}
	query.defaultTable = nil
	if len(query.Tables) > 0 {
		query.defaultTable = query.Tables[0]
	}
	query.keysMap = map[string]string{}
	fields := query.Fields
	for _, table := range query.Tables {
		fields = append(fields, table.Fields...)
	}
	for i := len(fields) - 1; i >= 0; i-- {
		field := fields[i]
		query.keysMap[field.Name] = field.GetKey()
	}
	query.active = true
}

// Close the query.
func (query *Query) Close() {
	query.active = false
}

// Select query sql SELECT.
func (query *Query) Select(db DB, selector sqlbuilder.Selector, args ...interface{}) ([]Map, error) {
	query.Open()
	qry := query.Selector.Merge(selector).SQL()
	return MapQuery(query.keysMap, db, qry, args...)
}

func (query *Query) execAction(c Context, onAction, defaultAction ActionFunc) (Result, error) {
	query.Open()
	if onAction == nil {
		return defaultAction(c)
	}
	return onAction(c)
}

func (query *Query) execDefaultAction(c Context) (Result, error) {
	query.Open()
	table := query.defaultTable
	if table == nil {
		return nil, errors.New("Query.Tables undefined")
	}
	c2 := Context{c.State, c.DB, query, table, c.Record, c.Field}
	switch c.State {
	case StateInsert:
		return table.execAction(c2, table.OnInsert, table.DefaultInsert)
	case StateUpdate:
		return table.execAction(c2, table.OnUpdate, table.DefaultUpdate)
	case StateDelete:
		return table.execAction(c2, table.OnDelete, table.DefaultDelete)
	default:
		return nil, errors.New("Invalid dataset state")
	}
}

// Insert execute sql INSERT INTO.
func (query *Query) Insert(db DB, record Map) (Result, error) {
	c := Context{StateInsert, db, query, nil, record, Field{}}
	return query.execAction(c, query.OnInsert, query.DefaultInsert)
}

// DefaultInsert is default Insert handler.
func (query *Query) DefaultInsert(c Context) (Result, error) {
	return query.execDefaultAction(c)
}

// Update execute sql UPDATE.
func (query *Query) Update(db DB, record Map) (Result, error) {
	c := Context{StateUpdate, db, query, nil, record, Field{}}
	return query.execAction(c, query.OnUpdate, query.DefaultUpdate)
}

// DefaultUpdate is default Update handler.
func (query *Query) DefaultUpdate(c Context) (Result, error) {
	return query.execDefaultAction(c)
}

// Delete execute sql DELETE;
func (query *Query) Delete(db DB, record Map) (Result, error) {
	c := Context{StateDelete, db, query, nil, record, Field{}}
	return query.execAction(c, query.OnDelete, query.DefaultDelete)
}

// DefaultDelete is default Delete handler.
func (query *Query) DefaultDelete(c Context) (Result, error) {
	return query.execDefaultAction(c)
}
