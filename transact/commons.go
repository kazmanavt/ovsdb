package transact

import (
	"github.com/kazmanavt/ovsdb/schema"
	"github.com/kazmanavt/ovsdb/types"
)

type operation interface {
	Name() string
	Validate(dSch *schema.DbSchema) error
}

type Result struct {
	Error   any            `json:"error,omitempty"`
	Details any            `json:"details,omitempty"`
	Uuid    types.UUIDType `json:"uuid,omitempty"`
	Rows    schema.Rows    `json:"rows,omitempty"`
	Count   int            `json:"count,omitempty"`
}
