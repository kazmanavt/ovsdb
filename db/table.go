package db

import (
	"encoding/json"
	"github.com/kazmanavt/ovsdb/monitor"
	"github.com/kazmanavt/ovsdb/schema"
	"github.com/kazmanavt/ovsdb/types"
	"sync"
)

type tableImpl struct {
	name   string
	sch    *schema.TableSchema
	cNames []string
	mu sync.RWMutex
	rows   map[string]schema.Row
}

func (t *tableImpl) findRecord(where []types.Condition) []string {
	t.mu.RLock()
	defer t.mu.RUnlock()
	var result []string
	for uuid, row := range t.rows {
		if where == nil || len(where) == 0 || row.Match(where) {
			result = append(result, uuid)
		}
	}
	return result
}


// apply updates
func (t *tableImpl) update2(upd2 monitor.TableUpdate2) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	for uuid, rowUpd2 := range upd2 {
		if rowUpd2.Initial != nil {
			row := t.sch.NewRow()
			if err := json.Unmarshal(rowUpd2.Initial, &row); err != nil {
				return err
			}
			t.rows[uuid] = row
		}
		if rowUpd2.Insert != nil {
			row := t.sch.NewRow()
			if err := json.Unmarshal(rowUpd2.Insert, &row); err != nil {
				return err
			}
			t.rows[uuid] = row
		}
		if rowUpd2.Delete != nil {
			delete(t.rows, uuid)
		}
		if rowUpd2.Modify != nil {
			row := t.sch.NewRow()
			if err := json.Unmarshal(rowUpd2.Modify, &row); err != nil {
				return err
			}
			if err := t.rows[uuid].Update2(row); err != nil {
				return err
			}
		}
	}
	return nil
}
