package monitor

import (
	"encoding/json"
	"fmt"
	"github.com/kazmanavt/ovsdb/schema"
)

type RawRowUpdate2 struct {
	Initial json.RawMessage `json:"initial,omitempty"`
	Insert  json.RawMessage `json:"insert,omitempty"`
	Delete  json.RawMessage `json:"delete,omitempty"`
	Modify  json.RawMessage `json:"modify,omitempty"`
}

type RawTableUpdate2 map[string]RawRowUpdate2
type RawTableSetUpdate2 map[string]RawTableUpdate2

func TableSetUpdateFromRaw2(dSch *schema.DbSchema, upd RawTableSetUpdate2) (TableSetUpdate2, error) {
	res := make(TableSetUpdate2)
	for tableName, rows := range upd {
		res[tableName] = make(TableUpdate2)
		tSch, ok := dSch.Tables[tableName]
		if !ok {
			return nil, fmt.Errorf("table %s not found in schema", tableName)
		}
		for rowId, row := range rows {
			_row := tSch.NewRow()
			switch {
			case row.Insert != nil:
				if err := _row.UnmarshalJSON(row.Insert); err == nil {
					res[tableName][rowId] = RowUpdate2{Insert: _row}
					continue
				}
			case row.Delete != nil:
				if err := _row.UnmarshalJSON(row.Delete); err == nil {
					res[tableName][rowId] = RowUpdate2{Delete: _row}
					continue
				}
			case row.Modify != nil:
				if err := _row.UnmarshalJSON(row.Modify); err == nil {
					res[tableName][rowId] = RowUpdate2{Modify: _row}
					continue
				}
			case row.Initial != nil:
				if err := _row.UnmarshalJSON(row.Initial); err == nil {
					res[tableName][rowId] = RowUpdate2{Initial: _row}
					continue
				}
			}
		}
	}
	return res, nil
}
