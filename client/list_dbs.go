package client

import (
	"context"
	"encoding/json"
	"log/slog"
)

func (c *Client) ListDbs(ctx context.Context) ([]string, error) {
	resp, err := c.jConn.Call(ctx, "list_dbs")
	if err != nil {
		return nil, err
	}

	var dbs []string
	if err := json.Unmarshal(resp.Result(), &dbs); err != nil {
		c.log.Debug("list dbs: fail unmarshal response",
			slog.String("raw", string(resp.Result())),
			slog.String("error", err.Error()))
		return nil, err
	}
	return dbs, nil
}
