package client

import (
	"context"
	"encoding/json"
	"fmt"
	jrpc "github.com/kazmanavt/jsonrpc/v1"
	"github.com/kazmanavt/ovsdb/monitor"
	"log/slog"
	"sync"
	"time"
)

const (
	defaultKeepAlivePeriod  = 30 * time.Second
	defaultKeepAliveTimeout = 5 * time.Second
)

type monitorItem struct {
	db          string
	lastTxnId   string
	monName     string
	initialReqs monitor.GenericMonReqSet
	renewReqs   monitor.GenericMonReqSet
	updChan2    chan<- monitor.TableSetUpdate2
	updChan     chan<- monitor.TableSetUpdate
}

type Client struct {
	log, jLog        *slog.Logger
	network, address string
	jConn            jrpc.Connection
	//ctx              context.Context
	//cancel           context.CancelFunc
	closed bool

	monitors map[string]*monitorItem
	monMu    sync.RWMutex

	keepAlivePeriod  time.Duration
	keepAliveTimeout time.Duration
}

func NewClient(network, addr string, opts ...ClientOpt) *Client {
	c := Client{
		network:          network,
		address:          addr,
		log:              slog.Default(),
		jLog:             slog.Default(),
		monitors:         make(map[string]*monitorItem),
		keepAlivePeriod:  defaultKeepAlivePeriod,
		keepAliveTimeout: defaultKeepAliveTimeout,
	}
	for _, opt := range opts {
		opt(&c)
	}

	c.connect()
	go c.loop()
	return &c
}

func (c *Client) keepAlive() {
	jConn := c.jConn
	seq := 0
	for {
		select {
		case <-jConn.Done():
			return
		case <-time.After(c.keepAlivePeriod):
			ctx, _ := context.WithTimeout(context.Background(), c.keepAliveTimeout)
			msg := fmt.Sprintf("keep alive %d", seq)
			resp, err := jConn.Call(ctx, "echo", msg)
			if err != nil {
				c.log.Warn("fail to send keep alive", slog.Any("local error", err))
				jConn.Close()
				return
			}
			if resp.Error() != nil {
				c.log.Warn("fail to send keep alive", slog.Any("remote error", resp.Error()))
				jConn.Close()
				return
			}

			var echoed []string
			if err := json.Unmarshal(resp.Result(), &echoed); err != nil {
				c.log.Warn("fail to send keep alive", slog.String("unmarshal error", err.Error()))
				jConn.Close()
				return
			}
			if len(echoed) != 1 {
				c.log.Warn("fail to send keep alive", slog.String("unexpected response length", string(resp.Result())))
				jConn.Close()
				return
			}
			if msg != echoed[0] {
				c.log.Warn("fail to send keep alive", slog.String("unexpected response", string(resp.Result())))
				jConn.Close()
				return
			}
		}
	}
}
func (c *Client) loop() {
	for {
		select {
		case <-c.jConn.Done():
			if c.closed {
				return
			}
			c.connect()
		}
	}
}

func (c *Client) connect() {
	c.log.Debug("creating new connection",
		slog.String("net", c.network),
		slog.String("addr", c.address))
	for {
		jConn, err := jrpc.NewClient(c.network, c.address, c.jLog)
		if err != nil {
			c.log.Warn("fail to connect to server", slog.Any("error", err))
			time.Sleep(1 * time.Second)
			continue
		}
		c.log.Debug("connected to server", slog.String("addr", c.network+"://"+c.address))

		if err := jConn.HandleCall("echo", c.echoHandler()); err != nil {
			c.log.Warn("fail to setup call echo handler", slog.Any("error", err))
			_ = jConn.Close()
			continue
		}
		if err = jConn.HandleNotification("update3", c.updates3Dispatcher()); err != nil {
			c.log.Warn("fail to setup update3 handler", slog.Any("error", err))
			_ = jConn.Close()
			continue
		}

		err = func() error {
			c.monMu.RLock()
			defer c.monMu.RUnlock()
			c.jConn = jConn
			if err := c.restoreMonitors(); err != nil {
				c.log.Warn("fail to restore monitors", slog.Any("error", err))
				_ = jConn.Close()
				return err
			}
			return nil
		}()
		if err != nil {
			continue
		}
		break
	}

	go c.keepAlive()

	c.log.Debug("connection established")
}

func (c *Client) restoreMonitors() error {
	c.log.Debug("restoring monitors")
	// restore monitors
	c.monMu.RLock()
	for _, item := range c.monitors {
		ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
		switch {
		case item.updChan2 != nil && item.renewReqs != nil:
			res, err := c.callMonitorCondSince(ctx, item.db, item.monName, item.lastTxnId, item.initialReqs)
			if err != nil {
				return err
			}
			if !res.found {
				item.updChan2 <- res.update2
			}
			item.lastTxnId = res.lastTxnID
		case item.updChan2 != nil && item.renewReqs == nil:
			upd2, err := c.callMonitorCond(ctx, item.db, item.monName, item.initialReqs)
			if err != nil {
				return err
			}
			item.updChan2 <- upd2
		case item.updChan != nil:
			upd, err := c.callMonitor(ctx, item.db, item.monName, item.initialReqs)
			if err != nil {
				return err
			}
			item.updChan <- upd
		}
	}
	c.monMu.RUnlock()
	return nil

}

func (c *Client) echoHandler() func(p json.RawMessage) (json.RawMessage, error) {
	return func(p json.RawMessage) (json.RawMessage, error) {
		c.log.Debug("echo handler", slog.String("server req.params", string(p)))
		return p, nil
	}
}

func (c *Client) updates3Dispatcher() func(string, string, monitor.TableSetUpdate2) {
	return func(monName string, txnId string, upd monitor.TableSetUpdate2) {
		c.log.Debug("updates dispatcher")
		c.monMu.RLock()
		item, ok := c.monitors[monName]
		c.monMu.RUnlock()
		if !ok {
			c.log.Warn("updates dispatcher", slog.String("monitor not found", monName))
			return
		}
		item.lastTxnId = txnId
		select {
		case item.updChan2 <- upd:
		default:
		}
	}
}

func (c *Client) Close() error {
	c.closed = true
	err := c.jConn.Close()
	return err
}
