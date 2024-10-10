package ovsdb

import (
	"context"
	"encoding/json"
	"github.com/kazmanavt/ovsdb/monitor"
	"log/slog"
)

func (c *Client) MonitorCondSince(ctx context.Context, db string, monName string, since string, monReqs monitor.MonCondReqs) (*monitor.Updates3, <-chan monitor.Updates3, error) {
	tUpdates2, uChan, err := c._monitor(ctx, "monitor_cond_since", db, monName, &since, monReqs)
	if err != nil {
		return nil, nil, err
	}

	var upd monitor.Updates3
	parms := []any{&upd.Found, &upd.LastTxnID, &upd.Upd2}
	err = json.Unmarshal(tUpdates2, &parms)
	if err != nil {
		close(uChan)
		return nil, nil, err
	}

	tuChan := make(chan monitor.Updates3, 10)
	go func() {
		for nParms := range uChan {
			var upd monitor.Updates3
			parms := []interface{}{new(string), &upd.LastTxnID, &upd.Upd2}
			err = json.Unmarshal(nParms, &parms)
			if err != nil {
				c.log.Error("fail to unmarshal updates", slog.String("error", err.Error()))
				continue
			}
			tuChan <- upd
		}
		close(tuChan)
	}()
	return &upd, tuChan, nil
}
