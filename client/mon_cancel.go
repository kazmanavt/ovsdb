package client

import "context"

func (c *Client) CancelMonitor(ctx context.Context, monName string) error {
	_, err := c.jConn.Call(ctx, "monitor_cancel", monName)

	return err
}
