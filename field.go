package godac

import (
	"strings"
	"time"
	"unicode"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

// Field is sql database table's column.
type Field struct {
	Name        string            // Column name
	Key         string            // Map key or JSON name/key
	Title       string            // The caption of column, used for display validation errors, etc...
	PrimaryKey  bool              // Is in primary key
	AutoInc     bool              // Is auto-increment
	ReadOnly    bool              // User cannot edit, it will be excluded on INSERT and UPDATE if Default or OnUpdate is nil.
	Default     interface{}       // Default value on INSERT
	OnUpdate    interface{}       // Value on UPDATE
	Validations []validation.Rule // Validation rules
}

// GetKey get real JSON Key or Map key, may be do naming conversion.
func (field Field) GetKey() string {
	if field.Key == "" {
		return convertName(field.Name)
	}
	return field.Key
}

// NamingConversionEnabled enable or disable naming conversion.
var NamingConversionEnabled = true

func convertName(name string) string {
	if NamingConversionEnabled {
		return camelCased(name)
	}
	return name
}

// UpperCasedWords define all uppercased words on camelCase naming conversion.
var UpperCasedWords = []string{"ID"}

func camelCased(s string) string {
	list := strings.Split(s, "_")
	for i, v := range list {
		if len(v) == 0 {
			continue
		}
		if i == 0 {
			if upperCasedWordsContains(v) {
				list[i] = strings.ToLower(v)
			} else {
				runes := []rune(v)
				runes[0] = unicode.ToLower(runes[0])
				list[i] = string(runes)
			}
		} else {
			if upperCasedWordsContains(v) {
				list[i] = strings.ToUpper(v)
			} else {
				list[i] = strings.Title(v)
			}
		}
	}
	return strings.Join(list, "")
}

func upperCasedWordsContains(s string) bool {
	s = strings.ToUpper(s)
	for _, v := range UpperCasedWords {
		if v == s {
			return true
		}
	}
	return false
}

// GetTitle get real title.
func (field Field) GetTitle() string {
	if field.Title == "" {
		list := strings.Split(field.Name, "_")
		for i, v := range list {
			if upperCasedWordsContains(v) {
				list[i] = strings.ToUpper(v)
			} else {
				list[i] = strings.Title(v)
			}
		}
		return strings.Join(list, " ")
	}
	return field.Title
}

// ValueFunc represents a get Default/OnUpdate value function.
type ValueFunc func() interface{}

// Now get current timestamp.
func Now() interface{} {
	return time.Now()
}

// GetDefault parse Default value.
func (field Field) GetDefault() interface{} {
	if caller, ok := field.Default.(ValueFunc); ok {
		return caller()
	}
	return field.Default
}

// GetOnUpdate parse OnUpdate value.
func (field Field) GetOnUpdate() interface{} {
	if caller, ok := field.OnUpdate.(ValueFunc); ok {
		return caller()
	}
	return field.OnUpdate
}
