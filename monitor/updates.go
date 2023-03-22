package monitor

import "encoding/json"

type RowUpdate struct {
	New json.RawMessage `json:"new,omitempty"`
	Old json.RawMessage `json:"old,omitempty"`
}

type TableUpdate map[string]RowUpdate

type Updates map[string]TableUpdate
