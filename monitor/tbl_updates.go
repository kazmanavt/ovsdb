package monitor

import (
	"github.com/kazmanavt/ovsdb/schema"
)

type RowUpdate struct {
	New schema.Row `json:"new,omitempty"`
	Old schema.Row `json:"old,omitempty"`
}

type TableUpdate map[string]RowUpdate

type TableSetUpdate map[string]TableUpdate
