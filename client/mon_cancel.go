package client

import "context"

func (c *Client) CancelMonitor(ctx context.Context, monName string) error {
	resp, err := c.jConn.Call(ctx, "monitor_cancel", monName)
	if err != nil {
		return err
	}
	if err := resp.Error(); err != nil {
		return err
	}

	return nil
}
