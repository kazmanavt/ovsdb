package monitor

import (
	"github.com/kazmanavt/ovsdb/v2/schema"
)

type RowUpdate2 struct {
	Initial schema.Row
	Insert  schema.Row
	Delete  schema.Row
	Modify  schema.Row
}
type TableUpdate2 map[string]RowUpdate2

type TableSetUpdate2 map[string]TableUpdate2
