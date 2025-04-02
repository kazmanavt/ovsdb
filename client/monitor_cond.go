package client

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/kazmanavt/ovsdb/monitor"
)

func u2FromJSON(c *Client, dbName string, data []byte) (monitor.TableSetUpdate2, error) {

	var upd2 monitor.RawTableSetUpdate2
	if err := json.Unmarshal(data, &upd2); err != nil {
		return nil, fmt.Errorf("unmarshal update2: %w", err)
	}

	c.schemasMu.RLock()
	dSch, ok := c.schemas[dbName]
	c.schemasMu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("db %s not found in schemas", dbName)
	}

	return monitor.TableSetUpdateFromRaw2(dSch, upd2)
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
	c.monMu.Lock()
	defer c.monMu.Unlock()

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
	c.monitors[monName] = &mon

	return upd2, tuChan, nil
}
