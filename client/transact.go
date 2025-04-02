package client

import (
	"context"
	"github.com/kazmanavt/ovsdb/v2/transact"
)

func (c *Client) Transact(ctx context.Context, db string, tr transact.Transaction) error {
	if err := tr.Validate(); err != nil {
		return err
	}
	args := []any{db}
	for _, op := range tr.Operations() {
		args = append(args, op)
	}
	resp, err := c.jConn.Call(ctx, "transact", args...)
	if err != nil {
		return err
	}
	if err := tr.DecodeResult(resp.GetResult()); err != nil {
		return err
	}
	return tr.Error()
}
