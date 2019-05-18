package godac

import v "github.com/go-ozzo/ozzo-validation"

// Field is sql database table's column.
type Field struct {
	Name         string
	Title        string
	InPrimaryKey bool
	IsAutoInc    bool
	ReadOnly     bool        // Excluded on INSERT and UPDATE if true, user cannot edit directly.
	Default      interface{} // Default value on INSERT
	OnUpdate     interface{} // Value on UPDATE
	Validations  []v.Rule    // validation rules
}
