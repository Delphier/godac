package godac

import (
	"godac/sqlbuilder"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

// Validation errors.
var (
	ErrUnique = validation.NewError("unique", "already exists")
	ErrIn     = validation.NewError("in", "not exists")
)

// ValidationRule is an extension of validation.Rule
type ValidationRule interface {
	SetContext(Context)
	validation.Rule
}

// Unique is a validation rule that checks if a value is unique in table.
var Unique = &uniqueRule{}

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

// In return a validation rule that check field value is in table.
func In(table *Table) *InRule {
	return &InRule{table: table}
}

// InRule is in validation rule object.
type InRule struct {
	context Context
	table   *Table
}

// SetContext implements ValidationRule.
func (rule *InRule) SetContext(c Context) {
	rule.context = c
}

// Validate implements validation.Rule.
func (rule *InRule) Validate(value interface{}) (err error) {
	count, err := rule.table.CountRecord(rule.context.DB, rule.context.Field, rule.context.Record, false, sqlbuilder.Select())
	if err != nil {
		return err
	}
	if count <= 0 {
		return ErrIn
	}
	return nil
}
