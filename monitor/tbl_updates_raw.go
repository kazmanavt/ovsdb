package monitor

import (
	"encoding/json"
	"fmt"
	"github.com/kazmanavt/ovsdb/v2/schema"
)

type RawRowUpdate struct {
	Old json.RawMessage `json:"old,omitempty"`
	New json.RawMessage `json:"new,omitempty"`
}

type RawTableUpdate map[string]RawRowUpdate
type RawTableSetUpdate map[string]RawTableUpdate

func TableSetUpdateFromRaw(dSch *schema.DbSchema, upd RawTableSetUpdate) (TableSetUpdate, error) {
	res := make(TableSetUpdate)
	for tableName, rows := range upd {
		res[tableName] = make(TableUpdate)
		tSch, ok := dSch.Tables[tableName]
		if !ok {
			return nil, fmt.Errorf("table %s not found in schema", tableName)
		}
		for rowId, row := range rows {
			_row := tSch.NewRow()
			switch {
			case row.Old != nil:
				if err := _row.UnmarshalJSON(row.Old); err == nil {
					res[tableName][rowId] = RowUpdate{Old: _row}
					continue
				}
			case row.New != nil:
				if err := _row.UnmarshalJSON(row.New); err == nil {
					res[tableName][rowId] = RowUpdate{New: _row}
					continue
				}
			}
		}
	}
	return res, nil
}
