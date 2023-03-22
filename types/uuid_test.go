package types

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestUUID_UnmarshalJSON(t *testing.T) {
	dataUUID := []byte(`["uuid", "8cc2eb5c-8e66-4554-af1d-8fa5b9321f99"]`)
	var u UUID
	t.Run("unmarshal UUID", func(t *testing.T) {
		assert.NoError(t, json.Unmarshal(dataUUID, &u))
		require.Equal(t, "8cc2eb5c-8e66-4554-af1d-8fa5b9321f99", string(u), "incorrect UUID")
	})
	t.Run("unmarshal UUID with wrong type", func(t *testing.T) {
		data := []byte(`["uud", "P"]`)
		assert.Error(t, json.Unmarshal(data, &u), "supposed to fail")
	})
	t.Run("unmarshal wrong data struct", func(t *testing.T) {
		data := []byte(`"uuid"`)
		assert.Error(t, json.Unmarshal(data, &u), "supposed to fail")
	})
}

func TestUUID_MarshalJSON(t *testing.T) {
	u := UUID("8cc2eb5c-8e66-4554-af1d-8fa5b9321f99")
	t.Run("marshal UUID", func(t *testing.T) {
		dataUUID, err := json.Marshal(u)
		require.NoError(t, err)
		require.Equal(t, `["uuid","8cc2eb5c-8e66-4554-af1d-8fa5b9321f99"]`, string(dataUUID), "incorrect JSON")
	})
}

func BenchmarkUUID_UnmarshalJSON(b *testing.B) {
	dataUUID := []byte(`["uuid", "8cc2eb5c-8e66-4554-af1d-8fa5b9321f99"]`)
	var u UUID
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = json.Unmarshal(dataUUID, &u)
	}
}
func BenchmarkUUID_MarshalJSON(b *testing.B) {
	u := UUID("8cc2eb5c-8e66-4554-af1d-8fa5b9321f99")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = json.Marshal(u)
	}
}
