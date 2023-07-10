package types

/*


import (
	"encoding/json"
	"fmt"
	"sync"
)

const namedUuidMark = "named-uuid"

// NamedUUID is a type for UUIDs, it is similar to UUID type, but it is used to represent
// UUIDs that are not yet known, e.g. when a row is inserted into a table and the UUID
// of the row is not known until the row is inserted. So it is possible to refer to the
// row inserted earlier in the same transaction using a NamedUUID.
// NamedUUID is meaningful only within the scope of a transaction.
type NamedUUID UUID

var namedUuidUnmarshalPool = sync.Pool{
	New: func() any {
		return make([]*string, 2)
	},
}

func (u *NamedUUID) UnmarshalJSON(data []byte) error {
	s := namedUuidUnmarshalPool.Get().([]*string)
	s[1] = (*string)(u)
	defer namedUuidUnmarshalPool.Put(s)
	if err := json.Unmarshal(data, &s); err != nil || *s[0] != namedUuidMark {
		return fmt.Errorf("invalid named-UUID: %w", err)
	}
	return nil
}

func (u NamedUUID) MarshalJSON() ([]byte, error) {
	// initial size is len(`["`+uuidMark+`","`)+len(u)+len(`"]`)
	b := make([]byte, 0, 11+len(u))
	b = append(b, `["`...)
	b = append(b, namedUuidMark...)
	b = append(b, `","`...)
	b = append(b, u...)
	b = append(b, `"]`...)
	return b, nil
}
*/
