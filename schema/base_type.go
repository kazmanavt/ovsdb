package schema

import (
	"encoding/json"
	"fmt"
	"github.com/kazmanavt/ovsdb/types"
	"reflect"
)

type BaseType struct {
	Type       string   `json:"type"`                 // type of the column
	EnumRaw    []any    `json:"enum,omitempty"`       // enum values of the column
	Enum       []any    `json:"-"`                    // enum values
	MinInteger *int     `json:"minInteger,omitempty"` // minimum integer value of the column
	MaxInteger *int     `json:"maxInteger,omitempty"` // maximum integer value of the column
	MinReal    *float64 `json:"minReal,omitempty"`    // minimum real value of the column
	MaxReal    *float64 `json:"maxReal,omitempty"`    // maximum real value of the column
	MinLength  *int     `json:"minLength,omitempty"`  // minimum length of the column
	MaxLength  *int     `json:"maxLength,omitempty"`  // maximum length of the column
	RefTable   *string  `json:"refTable,omitempty"`   // reference table of the column
	RefType    *string  `json:"refType,omitempty"`    // reference type of the column ("strong" or "weak")
}

func (bt *BaseType) UnmarshalJSON(data []byte) error {
	if err := json.Unmarshal(data, &bt.Type); err != nil {
		type BT BaseType
		if err := json.Unmarshal(data, (*BT)(bt)); err != nil {
			return fmt.Errorf("fail to unmarshal BaseType: %w", err)
		}
	}
	if bt.EnumRaw != nil {
		if mark, ok := bt.EnumRaw[0].(string); !ok || mark != "set" {
			return fmt.Errorf("invalid enum type: %v", bt.EnumRaw)
		}
		if values, ok := bt.EnumRaw[1].([]any); !ok || len(values) == 0 {
			return fmt.Errorf("invalid enum type: %v", bt.EnumRaw)
		}
		bt.Enum = bt.EnumRaw[1].([]any)
	}
	return nil
}

func validateEnumValue[T types.AtomicType](val T, eValues []any) error {
	for _, e := range eValues {
		if val == e.(T) {
			return nil
		}
	}
	return fmt.Errorf("wrong Enum value <%v>", val)
}

func (bt *BaseType) ValidateValue(v any, checkConstraints bool) error {
	switch typedVal := v.(type) {
	case int:
		if bt.Type != "integer" {
			return fmt.Errorf("expect %s got Int", bt.Type)
		}
		if bt.Enum != nil {
			return validateEnumValue(typedVal, bt.Enum)
		}
		if checkConstraints && (bt.MinInteger != nil && typedVal < *bt.MinInteger) {
			return fmt.Errorf("value [%d] < min [%d]", typedVal, *bt.MinInteger)
		}
		if checkConstraints && (bt.MaxInteger != nil && typedVal > *bt.MaxInteger) {
			return fmt.Errorf("value [%d] > max [%d]", typedVal, *bt.MaxInteger)
		}
		return nil
	case float64:
		if bt.Type != "real" {
			return fmt.Errorf("expect %s got Real", bt.Type)
		}
		if bt.Enum != nil {
			return validateEnumValue(typedVal, bt.Enum)
		}
		if checkConstraints && (bt.MinReal != nil && typedVal < *bt.MinReal) {
			return fmt.Errorf("value [%f] < min [%f]", typedVal, *bt.MinReal)
		}
		if checkConstraints && (bt.MaxReal != nil && typedVal > *bt.MaxReal) {
			return fmt.Errorf("value [%f] > max [%f]", typedVal, *bt.MaxReal)
		}
		return nil
	case string:
		if bt.Type != "string" {
			return fmt.Errorf("expect %s got String", bt.Type)
		}
		if bt.Enum != nil {
			return validateEnumValue(typedVal, bt.Enum)
		}
		if checkConstraints && (bt.MinLength != nil && len(typedVal) < *bt.MinLength) {
			return fmt.Errorf("value len(%s) < min %d", typedVal, *bt.MinLength)
		}
		if checkConstraints && (bt.MaxLength != nil && len(typedVal) > *bt.MaxLength) {
			return fmt.Errorf("value len(%s) > max [%d]", typedVal, *bt.MaxLength)
		}
		return nil
	case bool:
		if bt.Type != "boolean" {
			return fmt.Errorf("expect %s got Bool", bt.Type)
		}
		if bt.Enum != nil {
			return validateEnumValue(typedVal, bt.Enum)
		}
		return nil
	case types.UUID:
		if bt.Type != "uuid" {
			return fmt.Errorf("iexpect %s got UUID", bt.Type)
		}
		if bt.Enum != nil {
			return validateEnumValue(typedVal, bt.Enum)
		}
		return nil
	case types.NamedUUID:
		if bt.Type != "uuid" {
			return fmt.Errorf("iexpect %s got UUID", bt.Type)
		}
		if bt.Enum != nil {
			return validateEnumValue(types.UUID(typedVal), bt.Enum)
		}
		return nil
	}
	return fmt.Errorf("unsupported type: %v (%#+v)", reflect.TypeOf(v).Name(), v)
}
