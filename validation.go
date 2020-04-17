package godac

import validation "github.com/go-ozzo/ozzo-validation/v4"

// Unique is a validation rule that checks if a value is unique in table.
var Unique = &uniqueRule{}

// ErrUnique is validation error.
var ErrUnique = validation.NewError("unique", "already exists")

// Context contains the environment information on Insert/Update.
type Context struct {
	DB       DB
	Table    *Table
	Field    Field
	Record   Map
	IsInsert bool
}

// ValidationRule is an extension of validation.Rule
type ValidationRule interface {
	SetContext(Context)
	validation.Rule
}

type uniqueRule struct {
	context Context
}

func (rule *uniqueRule) SetContext(c Context) {
	rule.context = c
}

func (rule *uniqueRule) Validate(value interface{}) (err error) {
	var count int64
	if rule.context.IsInsert {
		count, err = rule.context.Table.CountValue(rule.context.DB, rule.context.Field, value, "")
	} else {
		count, err = rule.context.Table.CountRecord(rule.context.DB, rule.context.Field, rule.context.Record, true, "")
	}
	if err != nil {
		return
	}
	if count > 0 {
		return ErrUnique
	}
	return nil
}
