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
