package ovsdb

import (
	"context"
	"encoding/json"
	"log/slog"
)

func (c *Client) ListDbs(ctx context.Context) ([]string, error) {
	r, err := c.Call(ctx, "list_dbs")
	if err != nil {
		return nil, err
	}

	var dbs []string
	if err := json.Unmarshal(r, &dbs); err != nil {
		c.log.Debug("list dbs: fail unmarshal response",
			slog.String("raw", string(r)),
			slog.String("error", err.Error()))
		return nil, err
	}
	return dbs, nil
}
