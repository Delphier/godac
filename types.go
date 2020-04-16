package godac

import "database/sql"

// DB is a interface of *sql.DB and *sql.Tx.
type DB interface {
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
	Exec(query string, args ...interface{}) (sql.Result, error)
}

// Map is a shortcut for map[string]interface{}, represents a database record.
type Map map[string]interface{}

// Result is an extension of sql.Result.
type Result interface {
	Record(refresh bool) (Map, error) // Get last Insert/Update record, set refresh is true to requery from database.
	sql.Result
}

type result struct {
	sqlResult sql.Result
	isInsert  bool
	db        DB
	table     *Table
	record    Map
}

func (r result) Record(refresh bool) (Map, error) {
	if !refresh {
		return r.record, nil
	}
	var record = Map{}
	for k, v := range r.record {
		record[k] = v
	}
	if r.isInsert && r.table.autoInc >= 0 && r.table.Fields[r.table.autoInc].PrimaryKey {
		id, err := r.sqlResult.LastInsertId()
		if err != nil {
			return nil, err
		}
		record[r.table.keys[r.table.autoInc]] = id
	}
	query, args, err := r.table.WherePrimaryKey(record)
	if err != nil {
		return nil, err
	}
	maps, err := r.table.Select(r.db, "WHERE "+query, args...)
	if err != nil || len(maps) == 0 {
		return nil, err
	}
	for k, v := range maps[0] {
		record[k] = v
	}
	return record, nil
}

func (r result) LastInsertId() (int64, error) {
	return r.sqlResult.LastInsertId()
}
func (r result) RowsAffected() (int64, error) {
	return r.sqlResult.RowsAffected()
}
