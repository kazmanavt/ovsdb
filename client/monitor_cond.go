package client

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/kazmanavt/ovsdb/monitor"
)

type rowUpdate2 struct {
	Initial json.RawMessage `json:"initial,omitempty"`
	Insert  json.RawMessage `json:"insert,omitempty"`
	Delete  json.RawMessage `json:"delete,omitempty"`
	Modify  json.RawMessage `json:"modify,omitempty"`
}
type rawTableSetUpdate2 map[string]map[string]rowUpdate2

func (c *Client) tableSetUpdateFromRaw2(dbName string, upd rawTableSetUpdate2) (monitor.TableSetUpdate2, error) {
	c.schemasMu.RLock()
	dSch, ok := c.schemas[dbName]
	c.schemasMu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("db %s not found in schemas", dbName)
	}

	res := make(monitor.TableSetUpdate2)
	for tableName, rows := range upd {
		res[tableName] = make(monitor.TableUpdate2)
		tSch, ok := dSch.Tables[tableName]
		if !ok {
			return nil, fmt.Errorf("table %s not found in schema", tableName)
		}
		for rowId, row := range rows {
			_row := tSch.NewRow()
			switch {
			case row.Insert != nil:
				if err := _row.UnmarshalJSON(row.Insert); err == nil {
					res[tableName][rowId] = monitor.RowUpdate2{Insert: _row}
					continue
				}
			case row.Delete != nil:
				if err := _row.UnmarshalJSON(row.Delete); err == nil {
					res[tableName][rowId] = monitor.RowUpdate2{Delete: _row}
					continue
				}
			case row.Modify != nil:
				if err := _row.UnmarshalJSON(row.Modify); err == nil {
					res[tableName][rowId] = monitor.RowUpdate2{Modify: _row}
					continue
				}
			case row.Initial != nil:
				if err := _row.UnmarshalJSON(row.Initial); err == nil {
					res[tableName][rowId] = monitor.RowUpdate2{Initial: _row}
					continue
				}
			}
		}
	}
	return res, nil
}

func u2FromJSON(c *Client, dbName string, data []byte) (monitor.TableSetUpdate2, error) {

	var upd2 rawTableSetUpdate2
	if err := json.Unmarshal(data, &upd2); err != nil {
		return nil, fmt.Errorf("unmarshal update2: %w", err)
	}

	return c.tableSetUpdateFromRaw2(dbName, upd2)
}

func (c *Client) callMonitorCond(ctx context.Context, db string, monName string, monReqs monitor.GenericMonReqSet) (monitor.TableSetUpdate2, error) {
	resp, err := c._monitor(ctx, "monitor_cond", db, monName, nil, monReqs)
	if err != nil {
		return nil, err
	}

	if err := resp.Error(); err != nil {
		return nil, fmt.Errorf("monitor_cond: %w", err)
	}

	upd2, err := u2FromJSON(c, db, resp.GetResult())
	if err != nil {
		return nil, err
	}

	return upd2, nil
}
func (c *Client) SetMonitorCond(ctx context.Context, db string, monName string, monReqs monitor.MonCondReqSet) (monitor.TableSetUpdate2, <-chan monitor.TableSetUpdate2, error) {
	upd2, err := c.callMonitorCond(ctx, db, monName, monReqs)
	if err != nil {
		return nil, nil, err
	}

	tuChan := make(chan monitor.TableSetUpdate2, 10)
	mon := monitorItem{
		db:          db,
		monName:     monName,
		initialReqs: monReqs,
		renewReqs:   nil,
		updChan2:    tuChan,
		updChan:     nil,
	}
	c.monMu.Lock()
	c.monitors[monName] = &mon
	c.monMu.Unlock()

	return upd2, tuChan, nil
}
