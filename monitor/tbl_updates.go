package monitor

import (
	"github.com/kazmanavt/ovsdb/v2/schema"
)

type RowUpdate struct {
	New schema.Row `json:"new,omitempty"`
	Old schema.Row `json:"old,omitempty"`
}

type TableUpdate map[string]RowUpdate

type TableSetUpdate map[string]TableUpdate
