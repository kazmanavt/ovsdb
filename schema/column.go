package schema

import (
	"fmt"
	"github.com/kazmanavt/ovsdb/types"
	"reflect"
	"strings"
)

// ColumnSchema represents a column in a table, as defined in the OVSDB schema (RFC7047)
// A JSON object with the following members:
//
//	"type": <type>                            required
//	"ephemeral": <boolean>                    optional
//	"mutable": <boolean>                      optional
type ColumnSchema struct {
	Name      string     `json:"-"`
	Type      ColumnType `json:"type"`                // type of the column
	Ephemeral bool       `json:"ephemeral,omitempty"` // true if the column is ephemeral
	Mutable   bool       `json:"mutable,omitempty"`   // true if the column is mutable
}

// ValidateCond validates the condition for a column
func (cs *ColumnSchema) ValidateCond(op string, value any) (err error) {
	ops := "includes!==excludes"
	kind := cs.Type.GetKind()
	// normalize value for Set[int|float64]
	if v, ok := value.(int); kind == "Set[integer]" && *cs.Type.Min == 0 && *cs.Type.Max.(*int) == 1 && ok {
		ops = "<=>=" + ops
		value = types.Set[int]{v}
	} else if v, ok := value.(float64); kind == "Set[real]" && *cs.Type.Min == 0 && *cs.Type.Max.(*int) == 1 && ok {
		ops = "<=>=" + ops
		value = types.Set[float64]{v}
	}
	// validate operation
	if !strings.Contains(ops, op) {
		return fmt.Errorf("column %q: invalid operation %s", cs.Name, op)
	}

	// validate value
	checkConstrant := byte(3)
	if op == "includes" && (strings.HasPrefix(kind, "Set") || strings.HasPrefix(kind, "Map")) {
		checkConstrant &= ^byte(1)
	}
	if op == "excludes" && (strings.HasPrefix(kind, "Set") || strings.HasPrefix(kind, "Map")) {
		checkConstrant &= ^byte(2)
	}
	return cs.ValidateValue(value, checkConstrant)
}

func (cs *ColumnSchema) ValidateMutation(op string, value any) error {
	//if namedSet, ok := value.(types.Set[types.NamedUUID]); ok {
	//	var uuidSet types.Set[types.UUID]
	//	for _, named := range namedSet {
	//		uuidSet = append(uuidSet, types.UUID(named))
	//	}
	//	value = uuidSet
	//} else if named, ok := value.(types.NamedUUID); ok {
	//	value = types.UUID(named)
	//}

	ops := ""
	kind := cs.Type.GetKind()
	rv := reflect.ValueOf(value)
	switch rv.Kind() {
	case reflect.Slice, reflect.Map:
		ops = "insertdelete"
	case reflect.Int:
		ops = "+=-=*=/=%="
	case reflect.Float64:
		ops = "+=-=*=/="
	default:
		return fmt.Errorf("invalid type %s for mutate operation", kind)
	}
	if !strings.Contains(ops, op) {
		return fmt.Errorf("column %q: invalid operation %s", cs.Name, op)
	}
	if (strings.Contains("Set[integer]", kind) && rv.Kind() == reflect.Int) ||
		(strings.Contains("Set[real]", kind) && rv.Kind() == reflect.Float64) {
		return cs.Type.Key.ValidateValue(value, false)
	}
	if strings.HasPrefix(kind, "Set") && op == "insert" {
		return cs.ValidateValue(value, 2)
	}
	if strings.HasPrefix(kind, "Set") && op == "delete" {
		return cs.ValidateValue(value, 0)
	}
	if strings.HasPrefix(kind, "Map") && op == "insert" {
		return cs.ValidateValue(value, 2)
	}
	if strings.HasPrefix(kind, "Map") && op == "delete" {
		// TODO: validation case if delete from map by keys only
		return cs.ValidateValue(value, 0)
	}
	return cs.ValidateValue(value, 0)
}

func (cs *ColumnSchema) ValidateValue(value any, checks ...byte) error {
	checkConstraints, checkKeysConstraints, checkValuesConstraints := byte(3), byte(3), byte(3)
	switch len(checks) {
	case 0:
	case 1:
		checkConstraints = checks[0]
	case 2:
		checkConstraints, checkKeysConstraints = checks[0], checks[1]
	default:
		checkConstraints, checkKeysConstraints, checkValuesConstraints = checks[0], checks[1], checks[2]
	}
	if strings.HasPrefix(cs.Type.GetKind(), "Map[") {
		rv := reflect.ValueOf(value)
		if rv.Kind() != reflect.Map || !types.IsMapType(value) {
			return fmt.Errorf("expect map got %s", reflect.TypeOf(value).String())
		}
		if checkConstraints&0x1 != 0 && rv.Len() < *cs.Type.Min {
			return fmt.Errorf("column type constraint violation %d < %d", rv.Len(), *cs.Type.Min)
		}
		if checkConstraints&0x2 != 0 && rv.Len() > *cs.Type.Max.(*int) {
			return fmt.Errorf("column type constraint violation %d > %d", rv.Len(), *cs.Type.Max.(*int))
		}
		for _, k := range rv.MapKeys() {
			if err := cs.Type.Key.ValidateValue(k.Interface(), checkKeysConstraints > 0); err != nil {
				return err
			}
			if err := cs.Type.Value.ValidateValue(rv.MapIndex(k).Interface(), checkValuesConstraints > 0); err != nil {
				return err
			}
		}
		return nil
	} else if strings.HasPrefix(cs.Type.GetKind(), "Set[") {
		rv := reflect.ValueOf(value)
		if rv.Kind() != reflect.Slice || !types.IsSetType(value) {
			return fmt.Errorf("expect slice got %s", reflect.TypeOf(value).Name())
		}
		if rv.Len() < *cs.Type.Min || rv.Len() > *cs.Type.Max.(*int) {
			return fmt.Errorf("column type constraint violation %d not in range [%d, %d]", rv.Len(), *cs.Type.Min, *cs.Type.Max.(*int))
		}
		for i := 0; i < rv.Len(); i++ {
			if err := cs.Type.Key.ValidateValue(rv.Index(i).Interface(), checkKeysConstraints > 0); err != nil {
				return err
			}
		}
		return nil
	} else {
		return cs.Type.Key.ValidateValue(value, checkConstraints > 0)
	}
}

func (cs *ColumnSchema) GetDefaultValue() any {
	switch cs.Type.GetKind() {
	case "integer":
		var v int
		return v
	case "real":
		var v float64
		return v
	case "string":
		var v string
		return v
	case "boolean":
		var v bool
		return v
	case "uuid":
		var v types.UUIDType
		return v
	case "Set[integer]":
		return types.Set[int]{}
	case "Set[real]":
		return types.Set[float64]{}
	case "Set[string]":
		return types.Set[string]{}
	case "Set[boolean]":
		return types.Set[bool]{}
	case "Set[uuid]":
		return types.Set[types.UUIDType]{}
	case "Map[integer, integer]":
		return types.Map[int, int]{}
	case "Map[integer, real]":
		return types.Map[int, float64]{}
	case "Map[integer, string]":
		return types.Map[int, string]{}
	case "Map[integer, boolean]":
		return types.Map[int, bool]{}
	case "Map[integer, uuid]":
		return types.Map[int, types.UUIDType]{}
	case "Map[real, integer]":
		return types.Map[float64, int]{}
	case "Map[real, real]":
		return types.Map[float64, float64]{}
	case "Map[real, string]":
		return types.Map[float64, string]{}
	case "Map[real, boolean]":
		return types.Map[float64, bool]{}
	case "Map[real, uuid]":
		return types.Map[float64, types.UUIDType]{}
	case "Map[string, integer]":
		return types.Map[string, int]{}
	case "Map[string, real]":
		return types.Map[string, float64]{}
	case "Map[string, string]":
		return types.Map[string, string]{}
	case "Map[string, boolean]":
		return types.Map[string, bool]{}
	case "Map[string, uuid]":
		return types.Map[string, types.UUIDType]{}
	case "Map[boolean, integer]":
		return types.Map[bool, int]{}
	case "Map[boolean, real]":
		return types.Map[bool, float64]{}
	case "Map[boolean, string]":
		return types.Map[bool, string]{}
	case "Map[boolean, boolean]":
		return types.Map[bool, bool]{}
	case "Map[boolean, uuid]":
		return types.Map[bool, types.UUIDType]{}
	case "Map[uuid, integer]":
		return types.Map[types.UUIDType, int]{}
	case "Map[uuid, real]":
		return types.Map[types.UUIDType, float64]{}
	case "Map[uuid, string]":
		return types.Map[types.UUIDType, string]{}
	case "Map[uuid, boolean]":
		return types.Map[types.UUIDType, bool]{}
	case "Map[uuid, uuid]":
		return types.Map[types.UUIDType, types.UUIDType]{}
	default:
		return nil
	}
}
