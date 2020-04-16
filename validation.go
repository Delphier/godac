package godac

import validation "github.com/go-ozzo/ozzo-validation/v4"

// Unique is a validation rule that checks if a value is unique in table.
var Unique = &uniqueRule{}

// ErrUnique is validation error.
var ErrUnique = validation.NewError("unique", "already exists")

// ValidationRule is an extension of validation.Rule
type ValidationRule interface {
	Init(db DB, table *Table, field Field, record Map, isInsert bool)
	validation.Rule
}

type uniqueRule struct {
	db       DB
	table    *Table
	field    Field
	record   Map
	isInsert bool
}

func (rule *uniqueRule) Init(db DB, table *Table, field Field, record Map, isInsert bool) {
	rule.db = db
	rule.table = table
	rule.field = field
	rule.record = record
	rule.isInsert = isInsert
}

func (rule *uniqueRule) Validate(value interface{}) (err error) {
	var count int64
	if rule.isInsert {
		count, err = rule.table.CountValue(rule.db, rule.field, value, "")
	} else {
		count, err = rule.table.CountRecord(rule.db, rule.field, rule.record, true, "")
	}
	if err != nil {
		return
	}
	if count > 0 {
		return ErrUnique
	}
	return nil
}
