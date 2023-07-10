package ovsdb

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/kazmanavt/ovsdb/types"
)

func (c *Client) Echo(ctx context.Context) error {
	UUID := types.GetNamedUUID()
	result, err := c.Call(ctx, "echo", UUID)
	if err != nil {
		return err
	}
	var expect []string
	err = json.Unmarshal(result, &expect)
	if err != nil {
		return fmt.Errorf("fail parse echo result: %v", result)
	}
	if len(expect) != 1 || expect[0] != UUID {
		return fmt.Errorf("unexpected echo result: %v", expect)
	}
	return nil
}
