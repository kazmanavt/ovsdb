package schema

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestColumnType_UnmarshalJSON(t *testing.T) {
	t.Run("atomic-type", func(t *testing.T) {
		mm := 1
		t.Run("int", func(t *testing.T) {
			var ct ColumnType
			err := ct.UnmarshalJSON([]byte(`"integer"`))
			assert.NoError(t, err)
			assert.Equal(t, ColumnType{kind: "integer", Key: BaseType{Type: "integer"}, Min: &mm, Max: &mm}, ct)
		})
		t.Run("string", func(t *testing.T) {
			var ct ColumnType
			err := ct.UnmarshalJSON([]byte(`"string"`))
			assert.NoError(t, err)
			assert.Equal(t, ColumnType{kind: "string", Key: BaseType{Type: "string"}, Min: &mm, Max: &mm}, ct)
		})
		t.Run("uuid", func(t *testing.T) {
			var ct ColumnType
			err := ct.UnmarshalJSON([]byte(`"uuid"`))
			assert.NoError(t, err)
			assert.Equal(t, ColumnType{kind: "uuid", Key: BaseType{Type: "uuid"}, Min: &mm, Max: &mm}, ct)
		})
	})
	t.Run("non atomic-type", func(t *testing.T) {
		mm := 1
		t.Run("int", func(t *testing.T) {
			var ct ColumnType
			err := ct.UnmarshalJSON([]byte(`{"key":{"type": "integer", "minInteger": -1, "maxInteger": 1987}}`))
			assert.NoError(t, err)
			min, max := -1, 1987
			val := ColumnType{
				kind: "integer",
				Min:  &mm,
				Max:  &mm,
				Key: BaseType{
					Type:       "integer",
					MaxInteger: &max,
					MinInteger: &min,
				},
			}
			assert.Equal(t, val, ct)
		})
		t.Run("set[int]", func(t *testing.T) {
			var ct ColumnType
			err := ct.UnmarshalJSON([]byte(`{"key": "integer", "min": 0}`))
			assert.NoError(t, err)
			min := 0
			val := ColumnType{
				kind: "Set[integer]",
				Min:  &min,
				Max:  &mm,
				Key: BaseType{
					Type: "integer",
				},
			}
			assert.Equal(t, val, ct)
		})
		t.Run("map[string, int]", func(t *testing.T) {
			var ct ColumnType
			err := ct.UnmarshalJSON([]byte(`{"min": 0, "key": "string", "value": "integer"}`))
			assert.NoError(t, err)
			min := 0
			val := ColumnType{
				kind: "Map[string, integer]",
				Min:  &min,
				Max:  &mm,
				Key: BaseType{
					Type: "string",
				},
				Value: &BaseType{
					Type: "integer",
				},
			}
			assert.Equal(t, val, ct)
		})
		t.Run("map(limit[4:unlimited])[string, int(limit[0:4095])]", func(t *testing.T) {
			var ct ColumnType
			err := ct.UnmarshalJSON([]byte(`{"min":4, "max":"unlimited", "key":"string", "value": {"type":"integer","minInteger":0,"maxInteger":4095}}`))
			assert.NoError(t, err)
			vMin, vMax, min, max := 0, 4095, 4, int(^uint(0)>>1)
			val := ColumnType{
				kind: "Map[string, integer]",
				Min:  &min,
				Max:  &max,
				Key: BaseType{
					Type: "string",
				},
				Value: &BaseType{
					Type:       "integer",
					MinInteger: &vMin,
					MaxInteger: &vMax,
				},
			}
			assert.Equal(t, val, ct)
		})
	})
}
