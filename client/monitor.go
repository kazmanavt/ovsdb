package client

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/kazmanavt/ovsdb/monitor"
)

type rowUpdate struct {
	Old json.RawMessage `json:"old,omitempty"`
	New json.RawMessage `json:"new,omitempty"`
}
type rawTableSetUpdate map[string]map[string]rowUpdate

func (c *Client) tableSetUpdateFromRaw(dbName string, upd rawTableSetUpdate) (monitor.TableSetUpdate, error) {
	c.schemasMu.RLock()
	dSch, ok := c.schemas[dbName]
	c.schemasMu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("db %s not found in schemas", dbName)
	}

	res := make(monitor.TableSetUpdate)
	for tableName, rows := range upd {
		res[tableName] = make(monitor.TableUpdate)
		tSch, ok := dSch.Tables[tableName]
		if !ok {
			return nil, fmt.Errorf("table %s not found in schema", tableName)
		}
		for rowId, row := range rows {
			_row := tSch.NewRow()
			switch {
			case row.Old != nil:
				if err := _row.UnmarshalJSON(row.Old); err == nil {
					res[tableName][rowId] = monitor.RowUpdate{Old: _row}
					continue
				}
			case row.New != nil:
				if err := _row.UnmarshalJSON(row.New); err == nil {
					res[tableName][rowId] = monitor.RowUpdate{New: _row}
					continue
				}
			}
		}
	}
	return res, nil
}

func fromJSON(c *Client, dbName string, data []byte) (monitor.TableSetUpdate, error) {
	var upd rawTableSetUpdate
	if err := json.Unmarshal(data, &upd); err != nil {
		return nil, fmt.Errorf("unmarshal update: %w", err)
	}

	return c.tableSetUpdateFromRaw(dbName, upd)
}

func (c *Client) callMonitor(ctx context.Context, db string, monName string, monReqs monitor.GenericMonReqSet) (monitor.TableSetUpdate, error) {
	resp, err := c._monitor(ctx, "monitor", db, monName, nil, monReqs)
	if err != nil {
		return nil, err
	}

	if err := resp.Error(); err != nil {
		return nil, fmt.Errorf("monitor: %w", err)
	}

	upd, err := fromJSON(c, db, resp.GetResult())
	if err != nil {
		return nil, err
	}

	return upd, nil
}
func (c *Client) SetMonitor(ctx context.Context, db string, monName string, monReqs monitor.MonReqSet) (monitor.TableSetUpdate, <-chan monitor.TableSetUpdate, error) {
	upd, err := c.callMonitor(ctx, db, monName, monReqs)
	if err != nil {
		return nil, nil, err
	}

	tuChan := make(chan monitor.TableSetUpdate, 10)
	mon := monitorItem{
		db:          db,
		monName:     monName,
		initialReqs: monReqs,
		renewReqs:   nil,
		updChan2:    nil,
		updChan:     tuChan,
	}
	c.monMu.Lock()
	c.monitors[monName] = &mon
	c.monMu.Unlock()

	return upd, tuChan, nil
}
