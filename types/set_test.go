package types

import (
	_ "embed"
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

//go:embed testdata/set_0.json
var dataSet0 []byte

//go:embed testdata/set_1.json
var dataSet1 []byte

//go:embed testdata/set_N.json
var dataSetN []byte

func TestSet_UnmarshalJSON(t *testing.T) {
	t.Run("unmarshal empty set", func(t *testing.T) {
		var s Set[UUIDType]
		assert.NoError(t, json.Unmarshal(dataSet0, &s))
		require.Equal(t, Set[UUIDType]{}, s, "not empty set")
	})
	t.Run("unmarshal set with one element", func(t *testing.T) {
		var s Set[UUIDType]
		assert.NoError(t, json.Unmarshal(dataSet1, &s))
		require.Equal(t, Set[UUIDType]{UUID("8f5949cf-53d1-479b-af14-e44959d98967")}, s, "incorrect JSON")
	})
	t.Run("unmarshal set with N elements", func(t *testing.T) {
		var s Set[UUIDType]
		assert.NoError(t, json.Unmarshal(dataSetN, &s))
		require.ElementsMatch(t, Set[UUIDType]{
			UUID("88513ba2-70b0-49d3-9126-125e904cdd4d"),
			UUID("187deb8e-13c7-49b9-a785-8e3c7e7abdae"),
			UUID("91a503e4-01b6-479e-b8ed-c3f576e7ee40"),
		}, s, "incorrect JSON")
	})
}

func TestSet_MarshalJSON(t *testing.T) {
	t.Run("marshal empty set", func(t *testing.T) {
		s := Set[UUIDType]{}
		data, err := json.Marshal(s)
		require.NoError(t, err, "no error")
		require.JSONEq(t, `["set",[]]`, string(data), "incorrect JSON")
	})
	t.Run("marshal set with one element", func(t *testing.T) {
		s := Set[UUIDType]{UUID("8cc2eb5c-8e66-4554-af1d-8fa5b9321f99")}
		data, err := json.Marshal(s)
		require.NoError(t, err, "no error")
		require.JSONEq(t, `["uuid","8cc2eb5c-8e66-4554-af1d-8fa5b9321f99"]`, string(data), "incorrect JSON")
	})
	t.Run("marshal set with N elements", func(t *testing.T) {
		s := Set[UUIDType]{
			UUID("8cc2eb5c-8e66-4554-af1d-8fa5b9321f99"),
			UUID("187deb8e-13c7-49b9-a785-8e3c7e7abdae"),
			UUID("91a503e4-01b6-479e-b8ed-c3f576e7ee40"),
		}
		data, err := json.Marshal(s)
		require.NoError(t, err, "no error")
		require.JSONEq(t,
			`["set",[["uuid","8cc2eb5c-8e66-4554-af1d-8fa5b9321f99"],`+
				`["uuid","187deb8e-13c7-49b9-a785-8e3c7e7abdae"],`+
				`["uuid","91a503e4-01b6-479e-b8ed-c3f576e7ee40"]]]`,
			string(data), "incorrect JSON",
		)
	})
}

func BenchmarkSet_UnmarshalJSON(b *testing.B) {
	b.Run("unmarshal empty set", func(b *testing.B) {
		var s Set[UUIDType]
		for i := 0; i < b.N; i++ {
			_ = json.Unmarshal(dataSet0, &s)
		}
	})
	b.Run("unmarshal set with one element", func(b *testing.B) {
		var s Set[UUIDType]
		for i := 0; i < b.N; i++ {
			_ = json.Unmarshal(dataSet1, &s)
		}
	})
	b.Run("unmarshal set with N elements", func(b *testing.B) {
		var s Set[UUIDType]
		for i := 0; i < b.N; i++ {
			_ = json.Unmarshal(dataSetN, &s)
		}
	})
}

func BenchmarkSet_MarshalJSON(b *testing.B) {
	b.Run("marshal empty set", func(b *testing.B) {
		s := Set[UUIDType]{}
		for i := 0; i < b.N; i++ {
			_, _ = json.Marshal(s)
		}
	})
	b.Run("marshal set with one element", func(b *testing.B) {
		s := Set[UUIDType]{UUID("8cc2eb5c-8e66-4554-af1d-8fa5b9321f99")}
		for i := 0; i < b.N; i++ {
			_, _ = json.Marshal(s)
		}
	})
	b.Run("marshal set with N elements", func(b *testing.B) {
		s := Set[UUIDType]{
			UUID("8cc2eb5c-8e66-4554-af1d-8fa5b9321f99"),
			UUID("187deb8e-13c7-49b9-a785-8e3c7e7abdae"),
			UUID("91a503e4-01b6-479e-b8ed-c3f576e7ee40"),
			UUID("88513ba2-70b0-49d3-9126-125e904cdd4d"),
			UUID("8f5949cf-53d1-479b-af14-e44959d98967"),
			UUID("8cc2eb5c-8e66-4554-af1d-8fa5b9321f99"),
			UUID("187deb8e-13c7-49b9-a785-8e3c7e7abdae"),
			UUID("91a503e4-01b6-479e-b8ed-c3f576e7ee40"),
			UUID("88513ba2-70b0-49d3-9126-125e904cdd4d"),
			UUID("8f5949cf-53d1-479b-af14-e44959d98967"),
		}
		for i := 0; i < b.N; i++ {
			_, _ = json.Marshal(s)
		}
	})
}

func TestSet_Update(t *testing.T) {
	s := Set[UUIDType]{UUID("uuid-123"), UUID("uuid-456")}
	s.Update2(Set[UUIDType]{UUID("uuid-456"), UUID("uuid-789")})
	assert.ElementsMatch(t, Set[UUIDType]{UUID("uuid-123"), UUID("uuid-789")}, s)

	r := map[string]any{"test": s}

	x := r["test"].(Updater2)
	r["test"], _ = x.Update2(Set[UUIDType]{UUID("uuid-456"), UUID("uuid-789")})
	assert.ElementsMatch(t, Set[UUIDType]{UUID("uuid-123"), UUID("uuid-456")}, r["test"])
}
