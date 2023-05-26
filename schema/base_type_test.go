package schema

import (
	"github.com/kazmanavt/ovsdb/types"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBaseType_UnmarshalJSON(t *testing.T) {
	t.Run("atomic-type", func(t *testing.T) {
		t.Run("int", func(t *testing.T) {
			var bt BaseType
			err := bt.UnmarshalJSON([]byte(`"integer"`))
			assert.NoError(t, err)
			assert.Equal(t, BaseType{Type: "integer"}, bt)
		})
		t.Run("string", func(t *testing.T) {
			var bt BaseType
			err := bt.UnmarshalJSON([]byte(`"string"`))
			assert.NoError(t, err)
			assert.Equal(t, BaseType{Type: "string"}, bt)
		})
		t.Run("uuid", func(t *testing.T) {
			var bt BaseType
			err := bt.UnmarshalJSON([]byte(`"uuid"`))
			assert.NoError(t, err)
			assert.Equal(t, BaseType{Type: "uuid"}, bt)
		})
	})
	t.Run("base-type", func(t *testing.T) {
		t.Run("int", func(t *testing.T) {
			var bt BaseType
			err := bt.UnmarshalJSON([]byte(`{"type": "int", "minInteger": -1, "maxInteger": 1987}`))
			assert.NoError(t, err)
			min, max := -1, 1987
			assert.Equal(t, BaseType{Type: "int", MaxInteger: &max, MinInteger: &min}, bt)
		})
		t.Run("enum", func(t *testing.T) {
			var bt BaseType
			err := bt.UnmarshalJSON([]byte(`{"type": "string", "enum": ["set", ["foo", "bar"]]} `))
			assert.NoError(t, err)
			assert.Equal(t, BaseType{Type: "string", EnumRaw: []any{"set", []any{"foo", "bar"}}, Enum: []any{"foo", "bar"}}, bt)
		})
	})
}

func TestBaseType_ValidateValue(t *testing.T) {
	t.Run("atomic-type", func(t *testing.T) {
		t.Run("int", func(t *testing.T) {
			var bt BaseType
			err := bt.UnmarshalJSON([]byte(`"integer"`))
			assert.NoError(t, err)
			assert.NoError(t, bt.ValidateValue(89, true))
		})
		t.Run("string", func(t *testing.T) {
			var bt BaseType
			err := bt.UnmarshalJSON([]byte(`"string"`))
			assert.NoError(t, err)
			assert.NoError(t, bt.ValidateValue("erere", true))
		})
		t.Run("uuid", func(t *testing.T) {
			var bt BaseType
			err := bt.UnmarshalJSON([]byte(`"uuid"`))
			assert.NoError(t, err)
			assert.NoError(t, bt.ValidateValue(types.UUID("erere"), true))
		})
	})
	t.Run("base-type", func(t *testing.T) {
		t.Run("int", func(t *testing.T) {
			var bt BaseType
			err := bt.UnmarshalJSON([]byte(`{"type": "integer", "minInteger": -1, "maxInteger": 1987}`))
			assert.NoError(t, err)
			assert.NoError(t, bt.ValidateValue(89, true))
		})
		t.Run("enum", func(t *testing.T) {
			var bt BaseType
			err := bt.UnmarshalJSON([]byte(`{"type": "string", "enum": ["set", ["foo", "bar"]]} `))
			assert.NoError(t, err)
			assert.NoError(t, bt.ValidateValue("foo", true))
		})
		t.Run("enum2", func(t *testing.T) {
			var bt BaseType
			err := bt.UnmarshalJSON([]byte(`{"type": "string", "enum": ["set", ["active-backup", "balance-slb", "balance-tcp"]]} `))
			assert.NoError(t, err)
			assert.NoError(t, bt.ValidateValue("active-backup", true))
		})
		t.Run("enum out of range", func(t *testing.T) {
			var bt BaseType
			err := bt.UnmarshalJSON([]byte(`{"type": "string", "enum": ["set", ["foo", "bar"]]} `))
			assert.NoError(t, err)
			assert.Error(t, bt.ValidateValue("baz", true))
		})
		t.Run("uuid", func(t *testing.T) {
			var bt BaseType
			err := bt.UnmarshalJSON([]byte(`{"type": "uuid", "refTable": "foo"} `))
			assert.NoError(t, err)
			assert.NoError(t, bt.ValidateValue(types.UUID("erere"), true))
		})
	})
}
