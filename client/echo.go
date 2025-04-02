package client

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/kazmanavt/ovsdb/types"
)

func (c *Client) Echo(ctx context.Context) error {
	UUID := types.NewNamedUUID()
	resp, err := c.jConn.Call(ctx, "echo", UUID)
	if err != nil {
		return err
	}
	var expect []string
	err = json.Unmarshal(resp.GetResult(), &expect)
	if err != nil {
		return fmt.Errorf("fail parse echo resp: %v", resp)
	}
	if len(expect) != 1 || expect[0] != UUID {
		return fmt.Errorf("unexpected echo resp: %v", expect)
	}
	return nil
}
