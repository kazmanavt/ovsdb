package client

import (
	"context"
	"encoding/json"
	"github.com/kazmanavt/ovsdb/v2/schema"
	"log/slog"
)

func (c *Client) GetSchema(ctx context.Context, db string) (*schema.DbSchema, error) {
	c.schemasMu.RLock()
	if sch, ok := c.schemas[db]; ok {
		c.schemasMu.RUnlock()
		return sch, nil
	}
	c.schemasMu.RUnlock()

	resp, err := c.jConn.Call(ctx, "get_schema", db)
	if err != nil {
		return nil, err
	}
	if err := resp.Error(); err != nil {
		return nil, err
	}
	var sch schema.DbSchema
	if err := json.Unmarshal(resp.GetResult(), &sch); err != nil {
		c.log.Debug("get schema: fail unmarshal response", slog.String("error", err.Error()))
		return nil, err
	}
	c.schemasMu.Lock()
	defer c.schemasMu.Unlock()
	c.schemas[db] = &sch

	return &sch, nil
}
