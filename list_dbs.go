package ovsdb

import (
	"context"
	"encoding/json"
	"go.uber.org/zap"
)

func (c *Client) ListDbs(ctx context.Context) ([]string, error) {
	r, err := c.Call(ctx, "list_dbs")
	if err != nil {
		return nil, err
	}

	var dbs []string
	if err := json.Unmarshal(r, &dbs); err != nil {
		c.log.Debugw("list dbs: fail unmarshal response", zap.Error(err), zap.Any("raw", r))
		return nil, err
	}
	return dbs, nil
}
