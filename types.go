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
	Record() (Map, error) // Get last Insert or Update record from database.
	sql.Result
}

type result struct {
	sqlResult sql.Result
	isInsert  bool
	db        DB
	table     *Table
	pk        Map
}

func (r result) Record() (Map, error) {
	if r.isInsert && r.table.autoInc >= 0 && r.table.Fields[r.table.autoInc].PrimaryKey {
		id, err := r.sqlResult.LastInsertId()
		if err != nil {
			return nil, err
		}
		r.pk[r.table.keys[r.table.autoInc]] = id
	}
	query, args, err := r.table.WherePrimaryKey(r.pk)
	if err != nil {
		return nil, err
	}
	maps, err := r.table.Select(r.db, "WHERE "+query, args...)
	if err != nil || len(maps) == 0 {
		return nil, err
	}
	return maps[0], nil
}

func (r result) LastInsertId() (int64, error) {
	return r.sqlResult.LastInsertId()
}
func (r result) RowsAffected() (int64, error) {
	return r.sqlResult.RowsAffected()
}
