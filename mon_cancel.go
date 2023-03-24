package ovsdb

import "context"

func (c *Client) CancelMonitor(ctx context.Context, monName string) error {
	_, err := c.Call(ctx, "monitor_cancel", monName)

	c.uhMu.Lock()
	defer c.uhMu.Unlock()
	if uChan, ok := c.updatesHandlers[monName]; ok {
		close(uChan)
		delete(c.updatesHandlers, monName)
	}
	return err
}
