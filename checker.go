package godac

import (
	"fmt"
	"godac/sqlbuilder"
)

// CheckerFunc is used for some checks.
type CheckerFunc func(Context) error

// Exists return a checker used for checking field value is used in the other table.
func Exists(field Field, table *Table) CheckerFunc {
	return func(c Context) error {
		count, err := table.CountRecord(c.DB, field, c.Record, false, sqlbuilder.Select())
		if err != nil {
			return err
		}
		if count > 0 {
			return ErrExists
		}
		return nil
	}
}

// Deleter to get a OnDelete function by checkers.
func Deleter(checkers ...CheckerFunc) ActionFunc {
	return func(c Context) (Result, error) {
		for _, checker := range checkers {
			if err := checker(c); err != nil {
				if e, ok := err.(Error); ok {
					return nil, NewError(fmt.Sprintf(ErrorFormatOnDelete, e.Error()), e.Code())
				}
				return nil, err
			}
		}
		return c.Table.DefaultDelete(c)
	}
}
