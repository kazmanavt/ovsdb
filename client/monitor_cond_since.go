package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/kazmanavt/ovsdb/monitor"
	"github.com/kazmanavt/ovsdb/types"
)

type m3Resp struct {
	found     bool
	lastTxnID string
	update2   monitor.TableSetUpdate2
}

func (m3r *m3Resp) fromJSON(c *Client, dbName string, data []byte) error {
	var params []json.RawMessage
	if err := json.Unmarshal(data, &params); err != nil {
		return err
	}
	if len(params) != 3 {
		return errors.New("wrong number of param in Update3 notification")
	}
	if err := json.Unmarshal(params[0], &m3r.found); err != nil {
		return fmt.Errorf("unmarshal found: %w", err)
	}
	if err := json.Unmarshal(params[1], &m3r.lastTxnID); err != nil {
		return fmt.Errorf("unmarshal lastTxnID: %w", err)
	}
	var err error
	m3r.update2, err = u2FromJSON(c, dbName, params[2])
	if err != nil {
		return fmt.Errorf("unmarshal update2: %w", err)
	}
	return nil
}

func (c *Client) callMonitorCondSince(ctx context.Context, db string, monName, lastTxnId string, monReqs monitor.GenericMonReqSet) (m3Resp, error) {
	resp, err := c._monitor(ctx, "monitor_cond_since", db, monName, &lastTxnId, monReqs)
	if err != nil {
		return m3Resp{}, err
	}

	if err := resp.Error(); err != nil {
		return m3Resp{}, fmt.Errorf("monitor_cond_since: %w", err)
	}

	var res m3Resp
	err = res.fromJSON(c, db, resp.GetResult())
	if err != nil {
		return m3Resp{}, err
	}

	return res, nil
}

func (c *Client) SetMonitorCondSince(ctx context.Context, db string, monName string, monReqs monitor.MonCondReqSet) (monitor.TableSetUpdate2, <-chan monitor.TableSetUpdate2, error) {

	res, err := c.callMonitorCondSince(ctx, db, monName, types.ZeroUUID, monReqs)
	if err != nil {
		return nil, nil, err
	}

	tuChan := make(chan monitor.TableSetUpdate2, 10)
	mon := monitorItem{
		db:          db,
		lastTxnId:   res.lastTxnID,
		monName:     monName,
		initialReqs: monReqs,
		renewReqs:   monReqs.WithoutInitial(),
		updChan2:    tuChan,
		updChan:     nil,
	}
	c.monMu.Lock()
	c.monitors[monName] = &mon
	c.monMu.Unlock()

	return res.update2, tuChan, nil
}
