package monitor

import "encoding/json"

type RowUpdate2 struct {
	Initial json.RawMessage `json:"initial,omitempty"`
	Insert  json.RawMessage `json:"insert,omitempty"`
	Delete  json.RawMessage `json:"delete,omitempty"`
	Modify  json.RawMessage `json:"modify,omitempty"`
}

type TableUpdate2 map[string]RowUpdate2

type Updates2 map[string]TableUpdate2
