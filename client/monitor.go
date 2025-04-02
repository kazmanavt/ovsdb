package client

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/kazmanavt/ovsdb/v2/monitor"
)

func fromJSON(c *Client, dbName string, data []byte) (monitor.TableSetUpdate, error) {
	var upd monitor.RawTableSetUpdate
	if err := json.Unmarshal(data, &upd); err != nil {
		return nil, fmt.Errorf("unmarshal update: %w", err)
	}

	c.schemasMu.RLock()
	dSch, ok := c.schemas[dbName]
	c.schemasMu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("db %s not found in schemas", dbName)
	}

	return monitor.TableSetUpdateFromRaw(dSch, upd)
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
	c.monMu.Lock()
	defer c.monMu.Unlock()

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
	c.monitors[monName] = &mon

	return upd, tuChan, nil
}
