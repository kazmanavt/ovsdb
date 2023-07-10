package types

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"time"
)

const uuidMark = "uuid"
const namedUuidMark = "named-uuid"

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
	if err := json.Unmarshal(data, &s); err != nil || (*s[0] != uuidMark && *s[0] != namedUuidMark) {
		return fmt.Errorf("invalid UUID: %w", err)
	}
	return nil
}

func (u UUID) MarshalJSON() ([]byte, error) {
	// initial size is len(`["`+uuidMark+`","`)+len(u)+len(`"]`)
	var b []byte
	if strings.HasPrefix(string(u), "__") {
		b = make([]byte, 0, 17+len(u))
		b = append(b, `["`...)
		b = append(b, namedUuidMark...)
		b = append(b, `","`...)
		b = append(b, u...)
		b = append(b, `"]`...)
	} else {
		b = make([]byte, 0, 11+len(u))
		b = append(b, `["`...)
		b = append(b, uuidMark...)
		b = append(b, `","`...)
		b = append(b, u...)
		b = append(b, `"]`...)
	}
	return b, nil
}

// GetNamedUUID returns the random generated UUIDv4 string with prefix "__"
func GetNamedUUID() string {
	field := make([]byte, 16)
	r.Read(field[:])
	// version 4
	field[6] = (field[6] & 0x0f) | (4 << 4)
	// rfc 4122
	field[8] = (field[8] & 0x3f) | 0x80

	return fmt.Sprintf("__%x_%x_%x_%x_%x", field[0:4], field[4:6], field[6:8], field[8:10], field[10:])
}

var r = rand.New(rand.NewSource(time.Now().UnixNano()))
