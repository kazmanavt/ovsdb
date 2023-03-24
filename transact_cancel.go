package ovsdb

import "context"

func (c *Client) CancelTransact(ctx context.Context, id string) error {
	err := c.Notify(ctx, "cancel", id)
	if err != nil {
		c.DropPending(id)
	}
	return err
}
