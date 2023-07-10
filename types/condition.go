package types

import (
	"encoding/json"
	"errors"
	"reflect"
)

type Condition interface {
	json.Marshaler
	json.Unmarshaler
	GetColumn() string
	GetOp() string
	GetValue() any
	Check(value any) bool
}

type conditionImpl[T BaseType] struct {
	column   string
	function string
	value    T
}

func (c conditionImpl[T]) MarshalJSON() ([]byte, error) {
	return json.Marshal([]any{c.column, c.function, c.value})
}

func (c *conditionImpl[T]) UnmarshalJSON(data []byte) error {
	v := []any{&c.column, &c.function, &c.value}
	err := json.Unmarshal(data, &v)
	if err != nil {
		return err
	}
	if len(v) != 3 {
		return errors.New("invalid condition: must be a 3 element array")
	}
	return nil
}
func (c *conditionImpl[T]) GetColumn() string {
	return c.column
}
func (c *conditionImpl[T]) GetOp() string {
	return c.function
}
func (c *conditionImpl[T]) GetValue() any {
	return c.value
}

// Check checks if the value matches the condition.
// The value must be a pointer to a value of the same type as the condition value.
// in case of slice of int or float64 of 0 or 1 element condition value may be of type int or float64
func (c *conditionImpl[T]) Check(value any) bool {
	cv := any(c.value)
	rv := reflect.ValueOf(value)
	switch cvt := cv.(type) {
	case int:
		if rv.Kind() == reflect.Slice {
			if rv.Len() != 1 {
				return false
			}
			return checkNum(c.function, cvt, rv.Index(0).Interface())
		}
		return checkNum(c.function, cvt, value)
	case float64:
		if rv.Kind() == reflect.Slice {
			if rv.Len() != 1 {
				return false
			}
			return checkNum(c.function, cvt, rv.Index(0).Interface())
		}
		return checkNum(c.function, cvt, value)
	case string:
		return checkComp(c.function, cvt, value)
	case bool:
		return checkComp(c.function, cvt, value)
	case UUIDType:
		return checkComp(c.function, cvt, value)
	}
	switch reflect.ValueOf(cv).Kind() {
	case reflect.Slice:
		return checkSlice(c.function, c.value, value)
	case reflect.Map:
		return checkMap(c.function, c.value, value.(T))
	}
	return false
}

func checkSlice(function string, cv, value any) bool {
	rcv := reflect.ValueOf(cv)
	rv := reflect.ValueOf(value)
	switch function {
	case "==":
		return reflect.DeepEqual(cv, value)
	case "!=":
		return !reflect.DeepEqual(cv, value)
	case "includes":
		for i := 0; i < rcv.Len(); i++ {
			found := false
			for j := 0; j < rv.Len(); j++ {
				if reflect.DeepEqual(rcv.Index(i).Interface(), rv.Index(j).Interface()) {
					found = true
					break
				}
			}
			if !found {
				return false
			}
		}
		return true
	case "excludes":
		for i := 0; i < rcv.Len(); i++ {
			for j := 0; j < rv.Len(); j++ {
				if reflect.DeepEqual(rcv.Index(i).Interface(), rv.Index(j).Interface()) {
					return false
				}
			}
		}
		return true
	}
	return false
}

func checkMap[T any](function string, cv T, value T) bool {
	switch function {
	case "==":
		return reflect.DeepEqual(cv, value)
	case "!=":
		return !reflect.DeepEqual(cv, value)
	case "includes":
		iter := reflect.ValueOf(cv).MapRange()
		for iter.Next() {
			v1 := reflect.ValueOf(value).MapIndex(iter.Key())
			if !v1.IsValid() {
				return false
			}
			if !reflect.DeepEqual(v1.Interface(), iter.Value().Interface()) {
				return false
			}
		}
		return true
	case "excludes":
		iter := reflect.ValueOf(cv).MapRange()
		for iter.Next() {
			v1 := reflect.ValueOf(value).MapIndex(iter.Key())
			if !v1.IsValid() {
				continue
			}
			if reflect.DeepEqual(v1.Interface(), iter.Value().Interface()) {
				return false
			}
		}
		return true
	}
	return false
}

func checkComp[T string | bool | UUIDType](function string, cv T, value any) bool {
	vt := value.(T)
	switch function {
	case "==", "includes":
		return vt == cv
	case "!=", "excludes":
		return vt != cv
	}
	return false
}

func checkNum[T int | float64](function string, cv T, value any) bool {
	vt := value.(T)
	switch function {
	case "<":
		return vt < cv
	case "<=":
		return vt <= cv
	case ">":
		return vt > cv
	case ">=":
		return vt >= cv
	case "==", "includes":
		return vt == cv
	case "!=", "excludes":
		return vt != cv
	}
	return false
}

func LessThan[T BaseType](column string, value T) Condition {
	return &conditionImpl[T]{column, "<", value}
}
func LessEqual[T BaseType](column string, value T) Condition {
	return &conditionImpl[T]{column, "<=", value}
}

func GreaterThan[T BaseType](column string, value T) Condition {
	return &conditionImpl[T]{column, ">", value}
}
func GreaterEqual[T BaseType](column string, value T) Condition {
	return &conditionImpl[T]{column, ">=", value}
}

func Equal[T BaseType](column string, value T) Condition {
	return &conditionImpl[T]{column, "==", value}
}

func NotEqual[T BaseType](column string, value T) Condition {
	return &conditionImpl[T]{column, "!=", value}
}

func Includes[T BaseType](column string, value T) Condition {
	return &conditionImpl[T]{column, "includes", value}
}

func Excludes[T BaseType](column string, value T) Condition {
	return &conditionImpl[T]{column, "excludes", value}
}
