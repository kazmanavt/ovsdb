package client

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/kazmanavt/ovsdb/monitor"
)

func (c *Client) callMonitor(ctx context.Context, db string, monName string, monReqs monitor.GenericMonReqSet) (monitor.TableSetUpdate, error) {
	resp, err := c._monitor(ctx, "monitor_cond", db, monName, nil, monReqs)
	if err != nil {
		return nil, err
	}

	if resp.Error() != nil {
		return nil, fmt.Errorf("monitor_cond: %s", resp.Error())
	}

	var upd monitor.TableSetUpdate
	err = json.Unmarshal(resp.Result(), &upd)
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
