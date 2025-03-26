package client

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/kazmanavt/ovsdb/monitor"
)

func (c *Client) callMonitorCond(ctx context.Context, db string, monName string, monReqs monitor.GenericMonReqSet) (monitor.TableSetUpdate2, error) {
	resp, err := c._monitor(ctx, "monitor_cond", db, monName, nil, monReqs)
	if err != nil {
		return nil, err
	}

	if resp.Error() != nil {
		return nil, fmt.Errorf("monitor_cond: %s", resp.Error())
	}

	var upd2 monitor.TableSetUpdate2
	err = json.Unmarshal(resp.Result(), &upd2)
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
