package transact

import (
	"github.com/kazmanavt/ovsdb/v2/schema"
	"github.com/kazmanavt/ovsdb/v2/types"
)

type operation interface {
	Name() string
	Validate(dSch *schema.DbSchema) error
}

type Result struct {
	Error   any         `json:"error,omitempty"`
	Details any         `json:"details,omitempty"`
	Uuid    types.UUID  `json:"uuid,omitempty"`
	Rows    schema.Rows `json:"rows,omitempty"`
	Count   int         `json:"count,omitempty"`
}
