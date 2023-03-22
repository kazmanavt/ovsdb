package ovsdb

import (
	"context"
	"encoding/json"
	"github.com/kazmanavt/ovsdb/monitor"
)

func (c *Client) _monitor(ctx context.Context, monMethod string, db string, monName string, since *string, monReqs monitor.GenericMonReqs) (json.RawMessage, chan json.RawMessage, error) {
	if err := monReqs.Validate(); err != nil {
		return nil, nil, err
	}
	var tableUpdate json.RawMessage
	var err error
	if since == nil {
		tableUpdate, err = c.Call(ctx, monMethod, db, monName, monReqs)
	} else {
		tableUpdate, err = c.Call(ctx, monMethod, db, monName, monReqs, since)
	}
	if err != nil {
		return nil, nil, err
	}

	uChan := make(chan json.RawMessage, 10)
	if !monReqs.HasUpdates() {
		close(uChan)
		return tableUpdate, uChan, nil
	}

	c.uhMu.Lock()
	c.updatesHandlers[monName] = uChan
	c.uhMu.Unlock()
	return tableUpdate, uChan, nil
}
