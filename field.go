package godac

import "time"
import v "github.com/go-ozzo/ozzo-validation/v3"

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

// DefaultFunc represents a get default value function.
type DefaultFunc func() interface{}

// CurrentTimestamp get current datetime.
func CurrentTimestamp() interface{} {
	return time.Now()
}

// GetDefault parse Default value.
func (field Field) GetDefault() interface{} {
	caller, ok := field.Default.(DefaultFunc)
	if ok {
		return caller()
	}
	return field.Default
}

// GetOnUpdate parse OnUpdate value.
func (field Field) GetOnUpdate() interface{} {
	caller, ok := field.OnUpdate.(DefaultFunc)
	if ok {
		return caller()
	}
	return field.OnUpdate
}
