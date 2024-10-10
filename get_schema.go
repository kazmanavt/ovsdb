package ovsdb

import (
	"context"
	"encoding/json"
	"github.com/kazmanavt/ovsdb/schema"
	"log/slog"
)

func (c *Client) GetSchema(ctx context.Context, db string) (*schema.DbSchema, error) {
	result, err := c.Call(ctx, "get_schema", db)
	if err != nil {
		return nil, err
	}
	var sch schema.DbSchema
	if err := json.Unmarshal(result, &sch); err != nil {
		c.log.Debug("get schema: fail unmarshal response", slog.String("error", err.Error()))
		return nil, err
	}
	return &sch, nil
}
