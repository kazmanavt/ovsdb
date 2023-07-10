package types

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestCondition_MarshalJSON(t *testing.T) {
	t.Run("LessThan[int]", func(t *testing.T) {
		c := LessThan("x", 1)
		require.IsType(t, &conditionImpl[int]{}, c, "incorrect type")
		data, err := json.Marshal(c)
		require.NoError(t, err, "no error")
		require.JSONEq(t, `["x","<",1]`, string(data), "incorrect JSON")
	})
	t.Run("LessEqual[int]", func(t *testing.T) {
		c := LessEqual("x", 1)
		require.IsType(t, &conditionImpl[int]{}, c, "incorrect type")
		data, err := json.Marshal(c)
		require.NoError(t, err, "no error")
		require.JSONEq(t, `["x", "<=", 1]`, string(data), "incorrect JSON")
	})
	t.Run("GreaterThan[int]", func(t *testing.T) {
		c := GreaterThan("x", 1)
		require.IsType(t, &conditionImpl[int]{}, c, "incorrect type")
		data, err := json.Marshal(c)
		require.NoError(t, err, "no error")
		require.JSONEq(t, `["x", ">", 1]`, string(data), "incorrect JSON")
	})
	t.Run("GreaterEqual[int]", func(t *testing.T) {
		c := GreaterEqual("x", 1)
		require.IsType(t, &conditionImpl[int]{}, c, "incorrect type")
		data, err := json.Marshal(c)
		require.NoError(t, err, "no error")
		require.JSONEq(t, `["x", ">=", 1]`, string(data), "incorrect JSON")
	})
	t.Run("Equal[int]", func(t *testing.T) {
		c := Equal("x", 1)
		require.IsType(t, &conditionImpl[int]{}, c, "incorrect type")
		data, err := json.Marshal(c)
		require.NoError(t, err, "no error")
		require.JSONEq(t, `["x", "==", 1]`, string(data), "incorrect JSON")
	})
	t.Run("NotEqual[int]", func(t *testing.T) {
		c := NotEqual("x", 1)
		require.IsType(t, &conditionImpl[int]{}, c, "incorrect type")
		data, err := json.Marshal(c)
		require.NoError(t, err, "no error")
		require.JSONEq(t, `["x", "!=", 1]`, string(data), "incorrect JSON")
	})
	t.Run("Includes[int]", func(t *testing.T) {
		c := Includes("x", 1)
		require.IsType(t, &conditionImpl[int]{}, c, "incorrect type")
		data, err := json.Marshal(c)
		require.NoError(t, err, "no error")
		require.JSONEq(t, `["x", "includes", 1]`, string(data), "incorrect JSON")
	})
	t.Run("Excludes[int]", func(t *testing.T) {
		c := Excludes("x", 1)
		require.IsType(t, &conditionImpl[int]{}, c, "incorrect type")
		data, err := json.Marshal(c)
		require.NoError(t, err, "no error")
		require.JSONEq(t, `["x", "excludes", 1]`, string(data), "incorrect JSON")
	})

	t.Run("Equal[string]", func(t *testing.T) {
		c := Equal("x", "y")
		require.IsType(t, &conditionImpl[string]{}, c, "incorrect type")
		data, err := json.Marshal(c)
		require.NoError(t, err, "no error")
		require.JSONEq(t, `["x", "==", "y"]`, string(data), "incorrect JSON")
	})

	t.Run("Equal[UUID]", func(t *testing.T) {
		c := Equal("x", UUID("my-uuid-1-2-3"))
		require.IsType(t, &conditionImpl[UUIDType]{}, c, "incorrect type")
		data, err := json.Marshal(c)
		require.NoError(t, err, "no error")
		require.JSONEq(t, `["x", "==", ["uuid", "my-uuid-1-2-3"]]`, string(data), "incorrect JSON")
	})
	t.Run("Equal[NamedUUID]", func(t *testing.T) {
		c := Equal("x", NamedUUID("my-uuid-1-2-3"))
		require.IsType(t, &conditionImpl[UUIDType]{}, c, "incorrect type")
		data, err := json.Marshal(c)
		require.NoError(t, err, "no error")
		require.JSONEq(t, `["x", "==", ["named-uuid", "my-uuid-1-2-3"]]`, string(data), "incorrect JSON")
	})
	t.Run("Equal[Set[UUID]]", func(t *testing.T) {
		c := Equal("x", Set[UUIDType]{UUID("my-uuid-1-2-3")})
		require.IsType(t, &conditionImpl[Set[UUIDType]]{}, c, "incorrect type")
		data, err := json.Marshal(c)
		require.NoError(t, err, "no error")
		require.JSONEq(t, `["x", "==", ["uuid", "my-uuid-1-2-3"]]`, string(data), "incorrect JSON")
	})
	t.Run("Equal[Map[NamedUUID]]", func(t *testing.T) {
		c := Equal("x", Map[UUIDType, float64]{NamedUUID("my-uuid-1-2-3"): 1.0})
		require.IsType(t, &conditionImpl[Map[UUIDType, float64]]{}, c, "incorrect type")
		data, err := json.Marshal(c)
		require.NoError(t, err, "no error")
		require.JSONEq(t, `["x", "==", ["map", [[["named-uuid", "my-uuid-1-2-3"], 1]] ] ]`, string(data), "incorrect JSON")
	})
}

func TestCondition_GetColumn(t *testing.T) {
	t.Run("LessThan[int]", func(t *testing.T) {
		c := LessThan("x", 1)
		require.Equal(t, "x", c.GetColumn(), "incorrect column")
	})
}

func Test_conditionImpl_UnmarshalJSON(t *testing.T) {
	t.Run("LessThan[int]", func(t *testing.T) {
		c := &conditionImpl[int]{}
		err := json.Unmarshal([]byte(`["x", "<", 1]`), c)
		require.NoError(t, err, "no error")
		require.Equal(t, "x", c.GetColumn(), "incorrect column")
		require.Equal(t, 1, c.value, "incorrect value")
		require.Equal(t, "<", c.function, "incorrect operator")
		require.Equal(t, LessThan("x", 1), c, "bad conversion")
	})
	t.Run("bad condition", func(t *testing.T) {
		t.Run("too much values", func(t *testing.T) {
			c := &conditionImpl[int]{}
			err := json.Unmarshal([]byte(`["x", "<", 1, 2]`), c)
			require.Error(t, err, "error expected")
		})
		t.Run("too few values", func(t *testing.T) {
			c := &conditionImpl[int]{}
			err := json.Unmarshal([]byte(`["x", "<"]`), c)
			require.Error(t, err, "error expected")
		})
		t.Run("bad operator", func(t *testing.T) {
			c := &conditionImpl[int]{}
			err := json.Unmarshal([]byte(`["x", "!", 1]`), c)
			require.NoError(t, err, "error expected")
		})
	})
}

func Test_conditionImpl_Check(t *testing.T) {
	testsOK := []struct {
		name  string
		cond  Condition
		value any
	}{
		{
			name:  "LessThan[int]",
			cond:  LessThan("x", 1),
			value: 0,
		},
		{
			name:  "Equal[UUID]",
			cond:  Equal("x", UUID("my-uuid-1-2-3")),
			value: UUID("my-uuid-1-2-3"),
		},
		{
			name:  "Includes[Set[UUID]]",
			cond:  Includes("x", Set[UUIDType]{UUID("my-uuid-1-2-3")}),
			value: Set[UUIDType]{UUID("other-567"), UUID("my-uuid-1-2-3")},
		},
	}
	for _, tt := range testsOK {
		t.Run(tt.name, func(t *testing.T) {
			assert.True(t, tt.cond.Check(tt.value))
		})
	}
}
