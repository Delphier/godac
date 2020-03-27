package godac

import (
	"strings"
	"time"
	"unicode"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

// Field is sql database table's column.
type Field struct {
	Name         string
	Key          string // JSON name/key or Map key.
	Title        string
	InPrimaryKey bool
	IsAutoInc    bool
	ReadOnly     bool              // Excluded on INSERT and UPDATE if true, user cannot edit directly.
	Default      interface{}       // Default value on INSERT
	OnUpdate     interface{}       // Value on UPDATE
	Validations  []validation.Rule // validation rules
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

// GetKey get real JSON Key or Map key, may be do naming conversion.
func (field Field) GetKey() string {
	if field.Key == "" {
		return keyNamingConversion(field.Name)
	}
	return field.Key
}

// NamingConversionEnabled enable or disable naming conversion.
var NamingConversionEnabled = true

func keyNamingConversion(name string) string {
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
