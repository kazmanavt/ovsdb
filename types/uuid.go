package types

import (
	"encoding/json"
	"fmt"
	"sync"
)

const uuidMark = "uuid"
const namedUuidMark = "named-uuid"

// UUID is a type for UUIDs
// in OVSDBs JSON representation UUIDs are represented as a two element array with
// the first element being the string "uuid" and the second element being the UUID string
// e.g. ["uuid", "8cc2eb5c-8e66-4554-af1d-8fa5b9321f99"]

func UUID(uuid string) UUIDType {
	return UUIDType{kind: uuidMark, uuid: uuid}
}

func NamedUUID(uuid string) UUIDType {
	return UUIDType{kind: namedUuidMark, uuid: uuid}
}

type UUIDType struct {
	kind string
	uuid string
}

var uuidUnmarshalPool = sync.Pool{
	New: func() any {
		return make([]*string, 2)
	},
}

func (u *UUIDType) UnmarshalJSON(data []byte) error {
	s := uuidUnmarshalPool.Get().([]*string)
	s[0] = &u.kind
	s[1] = &u.uuid
	defer uuidUnmarshalPool.Put(s)
	if err := json.Unmarshal(data, &s); err != nil || (u.kind != uuidMark && u.kind != namedUuidMark) {
		return fmt.Errorf("invalid UUID: %w", err)
	}
	return nil
}

func (u UUIDType) MarshalJSON() ([]byte, error) {
	// initial size is len(`["`+uuidMark+`","`)+len(u)+len(`"]`)
	b := make([]byte, 0, 7+len(u.kind)+len(u.uuid))
	b = append(b, `["`...)
	b = append(b, u.kind...)
	b = append(b, `","`...)
	b = append(b, u.uuid...)
	b = append(b, `"]`...)
	return b, nil
}
