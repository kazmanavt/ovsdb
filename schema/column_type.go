package schema

import (
	"encoding/json"
	"fmt"
	"reflect"
)

type ColumnType struct {
	kind  string
	Key   BaseType  `json:"key"`             // key type of the column
	Value *BaseType `json:"value,omitempty"` // value type of the column
	Min   *int      `json:"min,omitempty"`   // minimum value of the column
	Max   any       `json:"max,omitempty"`   // maximum value of the column
}

func (ct *ColumnType) UnmarshalJSON(data []byte) error {
	if err := json.Unmarshal(data, &ct.Key.Type); err != nil {
		type CT ColumnType
		if err = json.Unmarshal(data, (*CT)(ct)); err != nil {
			return fmt.Errorf("fail to unmarshal cType from '%s' : %w", string(data), err)
		}
	}
	if ct.Min == nil {
		ct.Min = new(int)
		*ct.Min = 1
	}
	switch v := ct.Max.(type) {
	case float64:
		ct.Max = new(int)
		*ct.Max.(*int) = int(v)
	case string:
		if v != "unlimited" {
			return &json.UnsupportedValueError{
				Value: reflect.ValueOf(ct.Max),
				Str:   "should be string 'unlimited' or integer value",
			}
		}
		ct.Max = new(int)
		*ct.Max.(*int) = int(^uint(0) >> 1)
	case nil:
		ct.Max = new(int)
		*ct.Max.(*int) = 1
	default:
		return &json.UnsupportedTypeError{
			Type: reflect.TypeOf(ct.Max),
		}
	}
	if *ct.Min == 1 && *ct.Max.(*int) == 1 {
		ct.kind = ct.Key.Type
	}
	if *ct.Min < *ct.Max.(*int) {
		if ct.Value == nil {
			ct.kind = fmt.Sprintf("Set[%s]", ct.Key.Type)
		} else {
			ct.kind = fmt.Sprintf("Map[%s, %s]", ct.Key.Type, ct.Value.Type)
		}
	}
	return nil
}

func (ct *ColumnType) GetKind() string {
	return ct.kind
}
