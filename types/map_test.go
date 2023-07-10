package types

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"testing"
)

var (
	//go:embed testdata/map_0.json
	dataMap0 []byte

	//go:embed testdata/map_N.json
	dataMapN []byte
)

func TestMap_UnmarshalJSON(t *testing.T) {
	t.Run("unmarshal empty map", func(t *testing.T) {
		var m Map[string, int]
		assert.NoError(t, json.Unmarshal(dataMap0, &m))
		assert.Len(t, m, 0, "not empty map")
		assert.Equal(t, Map[string, int]{}, m, "not empty map")
	})
	t.Run("unmarshal map with N elements", func(t *testing.T) {
		var m Map[string, int]
		assert.NoError(t, json.Unmarshal(dataMapN, &m))
		assert.Len(t, m, 13, "incorrect map size")
		assert.Equal(t, Map[string, int]{
			`collisions`:       0,
			`rx_bytes`:         617930686,
			"rx_crc_err":       0,
			"rx_dropped":       84,
			"rx_errors":        0,
			"rx_frame_err":     0,
			"rx_missed_errors": 0,
			"rx_over_err":      0,
			"rx_packets":       1455542,
			"tx_bytes":         193157010,
			"tx_dropped":       0,
			"tx_errors":        0,
			"tx_packets":       347280,
		}, m, "incorrect map")
	})
}

func TestMap_MarshalJSON(t *testing.T) {
	t.Run("marshal empty map", func(t *testing.T) {
		m := Map[string, float64]{}
		m["foo"] = 1.0
		delete(m, "foo")
		data, err := json.Marshal(m)
		assert.NoError(t, err)
		assert.JSONEq(t, `["map", []]`, string(data), "incorrect JSON")
	})
	t.Run("marshal Map[string]int with N elements", func(t *testing.T) {
		m := Map[string, int]{
			"foo": 1,
			"bar": 2,
		}
		data, err := json.Marshal(m)
		assert.NoError(t, err)
		m2 := Map[string, int]{}
		err = json.Unmarshal(data, &m2)
		assert.NoError(t, err)
		assert.Equal(t, m, m2, "incorrect marshaling map")
	})
	t.Run("marshal map[string]UUID with N elements", func(t *testing.T) {
		m := Map[string, UUIDType]{
			"foo": UUID("UUID-123"),
			"bar": UUID("UUID-456"),
		}
		data, err := json.Marshal(m)
		assert.NoError(t, err)
		m2 := Map[string, UUIDType]{}
		err = json.Unmarshal(data, &m2)
		assert.NoError(t, err)
		assert.Equal(t, m, m2, "incorrect marshaling map")
	})
}

func benchMapUnmarshall[K, V AtomicType](b *testing.B, data []byte, m *Map[K, V]) {
	for i := 0; i < b.N; i++ {
		_ = json.Unmarshal(data, m)
	}
}
func BenchmarkMap_UnmarshalJSON(b *testing.B) {
	var (
		mStringInt    = Map[string, int]{}
		mStringString = Map[string, string]{}
		mStringUUID   = Map[string, UUIDType]{}
		mUUIDUUID     = Map[UUIDType, UUIDType]{}
	)
	for i := 0; i < 100; i++ {
		mStringInt[fmt.Sprintf("foo-%d", rand.Int())] = rand.Int()
		mStringString[fmt.Sprintf("foo-%d", rand.Int())] = fmt.Sprintf("bar-%d", rand.Int())
		mStringUUID[fmt.Sprintf("foo-%d", rand.Int())] = UUID(fmt.Sprintf("uuid-%d-fg-%d", rand.Int(), rand.Int()))
		mUUIDUUID[UUID(fmt.Sprintf("uuid-%d-fg-%d", rand.Int(), rand.Int()))] = UUID(fmt.Sprintf("uuid-%d-fg-%d", rand.Int(), rand.Int()))
	}

	data, _ := json.Marshal(mStringInt)
	b.Run("Map[string]int", func(b *testing.B) {
		var m Map[string, int]
		benchMapUnmarshall(b, data, &m)
	})
}

func benchMapMarshall[K, V AtomicType](b *testing.B, m Map[K, V]) {
	for i := 0; i < b.N; i++ {
		j, _ := json.Marshal(m)
		_ = j
	}
}

func BenchmarkMap_MarshalJSON(b *testing.B) {
	var (
		mStringInt    = Map[string, int]{}
		mStringString = Map[string, string]{}
		mStringUUID   = Map[string, UUIDType]{}
		mUUIDUUID     = Map[UUIDType, UUIDType]{}
	)
	for i := 0; i < 100; i++ {
		mStringInt[fmt.Sprintf("foo-%d", rand.Int())] = rand.Int()
		mStringString[fmt.Sprintf("foo-%d", rand.Int())] = fmt.Sprintf("bar-%d", rand.Int())
		mStringUUID[fmt.Sprintf("foo-%d", rand.Int())] = UUID(fmt.Sprintf("uuid-%d-fg-%d", rand.Int(), rand.Int()))
		mUUIDUUID[UUID(fmt.Sprintf("uuid-%d-fg-%d", rand.Int(), rand.Int()))] = UUID(fmt.Sprintf("uuid-%d-fg-%d", rand.Int(), rand.Int()))
	}

	b.ResetTimer()

	b.Run("Map[string]int", func(b *testing.B) {
		benchMapMarshall(b, mStringInt)
	})

	b.Run("Map[string]string", func(b *testing.B) {
		benchMapMarshall(b, mStringString)
	})
	b.Run("Map[string]UUID", func(b *testing.B) {
		benchMapMarshall(b, mStringUUID)
	})
	b.Run("Map[UUID]UUID", func(b *testing.B) {
		benchMapMarshall(b, mUUIDUUID)
	})
}

func TestMap_Update(t *testing.T) {
	m := Map[string, int]{"foo": 1, "bar": 2}
	m.Update2(Map[string, int]{"foo": 1, "bar": 4, "qux": 5})
	assert.Equal(t, Map[string, int]{"bar": 4, "qux": 5}, m)
}
