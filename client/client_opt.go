package client

import (
	"log/slog"
	"time"
)

type ClientOpt func(c *Client)

func WithLogger(l *slog.Logger) ClientOpt {
	return func(c *Client) {
		c.log = l
	}
}

func WithJLogger(l *slog.Logger) ClientOpt {
	return func(c *Client) {
		c.jLog = l
	}
}

func WithKeepAlivePeriod(period time.Duration) ClientOpt {
	return func(c *Client) {
		c.keepAlivePeriod = period
	}
}

func WithKeepAliveTimeout(timeout time.Duration) ClientOpt {
	return func(c *Client) {
		c.keepAliveTimeout = timeout
	}
}
