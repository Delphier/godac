package godac

import (
	"godac/sqlbuilder"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

// Unique is a validation rule that checks if a value is unique in table.
var Unique = &uniqueRule{}

// ErrUnique is validation error.
var ErrUnique = validation.NewError("unique", "already exists")

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
	if rule.context.State == StateInsert {
		count, err = rule.context.Table.CountValue(rule.context.DB, rule.context.Field, value, sqlbuilder.Select())
	} else {
		count, err = rule.context.Table.CountRecord(rule.context.DB, rule.context.Field, rule.context.Record, true, sqlbuilder.Select())
	}
	if err != nil {
		return
	}
	if count > 0 {
		return ErrUnique
	}
	return nil
}
