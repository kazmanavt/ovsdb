package client

import (
	"context"
	jrpc "github.com/kazmanavt/jsonrpc/v2/jrpc1"
	"github.com/kazmanavt/ovsdb/v2/monitor"
)

func (c *Client) _monitor(ctx context.Context, monMethod string, db string, monName string, since *string, monReqs monitor.GenericMonReqSet) (jrpc.Response, error) {
	if err := monReqs.Validate(); err != nil {
		return nil, err
	}
	var resp jrpc.Response
	var err error
	if since == nil {
		resp, err = c.jConn.Call(ctx, monMethod, db, monName, monReqs)
	} else {
		resp, err = c.jConn.Call(ctx, monMethod, db, monName, monReqs, since)
	}
	if err != nil {
		return nil, err
	}

	if err := resp.Error(); err != nil {
		return nil, err
	}

	return resp, nil
}
