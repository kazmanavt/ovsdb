package types

import (
	"encoding/json"
	"fmt"
	"sync"
)

const uuidMark = "uuid"

// UUID is a type for UUIDs
// in OVSDBs JSON representation UUIDs are represented as a two element array with
// the first element being the string "uuid" and the second element being the UUID string
// e.g. ["uuid", "8cc2eb5c-8e66-4554-af1d-8fa5b9321f99"]
type UUID string

var uuidUnmarshalPool = sync.Pool{
	New: func() any {
		return make([]*string, 2)
	},
}

func (u *UUID) UnmarshalJSON(data []byte) error {
	s := uuidUnmarshalPool.Get().([]*string)
	s[1] = (*string)(u)
	defer uuidUnmarshalPool.Put(s)
	if err := json.Unmarshal(data, &s); err != nil || *s[0] != uuidMark {
		return fmt.Errorf("invalid UUID: %w", err)
	}
	return nil
}

func (u UUID) MarshalJSON() ([]byte, error) {
	// initial size is len(`["`+uuidMark+`","`)+len(u)+len(`"]`)
	b := make([]byte, 0, 11+len(u))
	b = append(b, `["`...)
	b = append(b, uuidMark...)
	b = append(b, `","`...)
	b = append(b, u...)
	b = append(b, `"]`...)
	return b, nil
}
