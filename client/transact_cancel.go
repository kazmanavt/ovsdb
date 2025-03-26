package client

import "context"

func (c *Client) CancelTransact(ctx context.Context, id string) error {
	err := c.jConn.Notify(ctx, "cancel", id)
	return err
}
