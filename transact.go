package ovsdb

import (
	"context"
	"github.com/kazmanavt/ovsdb/transact"
)

func (c *Client) Transact(ctx context.Context, db string, tr transact.Transaction) error {
	if err := tr.Validate(); err != nil {
		return err
	}
	args := []any{db}
	for _, op := range tr.Operations() {
		args = append(args, op)
	}
	result, err := c.Call(ctx, "transact", args...)
	if err != nil {
		return err
	}
	if err := tr.DecodeResult(result); err != nil {
		return err
	}
	return tr.Error()
}
func (c *Client) CancelTransact(ctx context.Context, id string) error {
	err := c.Notify(ctx, "cancel", id)
	if err != nil {
		c.DropPending(id)
	}
	return err
}
