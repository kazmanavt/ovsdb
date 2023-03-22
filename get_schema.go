package ovsdb

import (
	"context"
	"encoding/json"
	"github.com/kazmanavt/ovsdb/schema"
	"go.uber.org/zap"
)

func (c *Client) GetSchema(ctx context.Context, db string) (*schema.DbSchema, error) {
	result, err := c.Call(ctx, "get_schema", db)
	if err != nil {
		return nil, err
	}
	//sch := schema.DbSchema{
	//	Tables: map[string]schema.TableSchema{},
	//}
	var sch schema.DbSchema
	if err := json.Unmarshal(result, &sch); err != nil {
		c.log.Debugw("get schema: fail unmarshal response", zap.Error(err))
		return nil, err
	}
	return &sch, nil
}
