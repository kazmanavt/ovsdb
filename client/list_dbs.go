package client

import (
	"context"
	"encoding/json"
	"log/slog"
	"slices"
)

func (c *Client) ListDbs(ctx context.Context) ([]string, error) {
	c.dbsNamesMu.RLock()
	if c.dbsNames != nil {
		return slices.Clone(c.dbsNames), nil
	}
	c.dbsNamesMu.RUnlock()

	resp, err := c.jConn.Call(ctx, "list_dbs")
	if err != nil {
		return nil, err
	}
	if err := resp.Error(); err != nil {
		return nil, err
	}

	var dbs []string
	if err := json.Unmarshal(resp.GetResult(), &dbs); err != nil {
		c.log.Debug("list dbs: fail unmarshal response",
			slog.String("raw", string(resp.GetResult())),
			slog.String("error", err.Error()))
		return nil, err
	}
	c.dbsNamesMu.Lock()
	defer c.dbsNamesMu.Unlock()
	c.dbsNames = dbs

	return slices.Clone(dbs), nil
}
