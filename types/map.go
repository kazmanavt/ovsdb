package types

import (
	//"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
)

const (
	mapMark = "map"
)

// Map is a type for Maps
// in OVSDBs JSON representation Maps are represented as a two element array with
// the first element being the string "map" and the second element being an array of two element arrays
// e.g. ["map", [[["uuid", "8cc2eb5c-8e66-4554-af1d-8fa5b9321f99"], ["uuid", "8cc2eb5c-8e66-4554-af1d-8fa5b9321f99"]]]]
// or ["map", []] for an empty map
type Map[K AtomicType, V AtomicType] map[K]V

func (m *Map[K, V]) UnmarshalJSON(data []byte) error {
	var mark string
	body := [][2]any{}
	parts := [2]any{&mark, &body}
	if err := json.Unmarshal(data, &parts); err != nil || mark != mapMark {
		return fmt.Errorf("invalid map (no mark or err: %w)", err)
	}
	for i := range body {
		var key K
		var value V
		body[i] = [2]any{&key, &value}
	}
	if err := json.Unmarshal(data, &parts); err != nil {
		return err
	}
	_m := make(map[K]V)
	for i := range body {
		_m[*body[i][0].(*K)] = *body[i][1].(*V)
	}
	*m = _m
	return nil
}

var mapMarshalPool = sync.Pool{
	New: func() any {
		return [][2]any{}
	},
}

func (m Map[K, V]) MarshalJSON() ([]byte, error) {
	//b := make([][2]any, 0, len(m))
	b := mapMarshalPool.Get().([][2]any)
	if cap(b) < len(m) {
		b = make([][2]any, len(m))
	}
	i := 0
	for k, v := range m {
		b[i][0] = k
		b[i][1] = v
		i++
	}
	//b = b[:i]
	//b = append(b, [2]any{k, v})
	defer mapMarshalPool.Put(b)
	return json.Marshal([2]any{mapMark, b[:i]})
}

func (m *Map[K, V]) String() string {
	var sb strings.Builder
	sb.WriteString("{")
	d := ""
	for k, v := range *m {
		_, _ = fmt.Fprintf(&sb, "%s%v: %v", d, k, v)
		d = ", "
	}
	sb.WriteString("}")
	return sb.String()
}

func (m Map[K, V]) Update2(_m2 any) error {
	m2, ok := _m2.(Map[K, V])
	if !ok {
		return fmt.Errorf("invalid type for Map update2")
	}
	if m2 == nil {
		return nil
	}
	for k2, v2 := range m2 {
		found := false
		for k, v := range m {
			if k2 == k {
				found = true
				if v2 == v {
					delete(m, k)
				} else {
					(m)[k] = v2
				}
				break
			}
		}
		if !found {
			(m)[k2] = v2
		}
	}
	return nil
}
