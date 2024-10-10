package ovsdb

import (
	"context"
	"encoding/json"
	"github.com/kazmanavt/ovsdb/monitor"
	"log/slog"
)

func (c *Client) Monitor(ctx context.Context, db string, monName string, monReqs monitor.MonReqs) (monitor.Updates, <-chan monitor.Updates, error) {
	tUpdates, uChan, err := c._monitor(ctx, "monitor", db, monName, nil, monReqs)
	if err != nil {
		return nil, nil, err
	}

	var upd monitor.Updates
	err = json.Unmarshal(tUpdates, &upd)
	if err != nil {
		_ = c.CancelMonitor(nil, monName)
		return nil, nil, err
	}

	tuChan := make(chan monitor.Updates, 10)
	go func() {
		for nParms := range uChan {
			var upd monitor.Updates
			parms := []any{new(string), upd}
			err = json.Unmarshal(nParms, &parms)
			if err != nil {
				c.log.Error("fail to unmarshal updates", slog.String("error", err.Error()))
				continue
			}
			tuChan <- upd
		}
		close(tuChan)
	}()
	return upd, tuChan, nil
}
