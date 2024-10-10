package ovsdb

import (
	"context"
	"encoding/json"
	"github.com/kazmanavt/ovsdb/monitor"
	"log/slog"
)

func (c *Client) MonitorCond(ctx context.Context, db string, monName string, monReqs monitor.MonCondReqs) (monitor.Updates2, <-chan monitor.Updates2, error) {
	tUpdates2, uChan, err := c._monitor(ctx, "monitor_cond", db, monName, nil, monReqs)
	if err != nil {
		return nil, nil, err
	}

	var upd2 monitor.Updates2
	err = json.Unmarshal(tUpdates2, &upd2)
	if err != nil {
		close(uChan)
		return nil, nil, err
	}

	tuChan := make(chan monitor.Updates2, 10)
	go func() {
		for nParms := range uChan {
			var upd2 monitor.Updates2
			parms := []any{new(string), &upd2}
			err = json.Unmarshal(nParms, &parms)
			if err != nil {
				c.log.Error("fail to unmarshal updates", slog.String("error", err.Error()))
				continue
			}
			tuChan <- upd2
		}
		close(tuChan)
	}()
	return upd2, tuChan, nil
}
