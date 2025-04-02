package client

import (
	"context"
	"encoding/json"
	"fmt"
	jrpc "github.com/kazmanavt/jsonrpc/v1"
	"github.com/kazmanavt/ovsdb/monitor"
	"github.com/kazmanavt/ovsdb/schema"
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
	lock   sync.RWMutex
	closed bool

	monitors map[string]*monitorItem
	monMu    sync.RWMutex

	schemas   map[string]*schema.DbSchema
	schemasMu sync.RWMutex

	dbsNames   []string
	dbsNamesMu sync.RWMutex

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
		schemas:          make(map[string]*schema.DbSchema),
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
			if err := resp.Error(); err != nil {
				c.log.Warn("fail to send keep alive", slog.Any("remote error", err.Error()))
				jConn.Close()
				return
			}

			var echoed []string
			if err := json.Unmarshal(resp.GetResult(), &echoed); err != nil {
				c.log.Warn("fail to send keep alive", slog.String("unmarshal error", err.Error()))
				jConn.Close()
				return
			}
			if len(echoed) != 1 {
				c.log.Warn("fail to send keep alive", slog.String("unexpected response length", string(resp.GetResult())))
				jConn.Close()
				return
			}
			if msg != echoed[0] {
				c.log.Warn("fail to send keep alive", slog.String("unexpected response", string(resp.GetResult())))
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
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()
	for {
		jConn, err := jrpc.NewClient(c.network, c.address, c.jLog)
		if err != nil {
			c.log.Warn("fail to connect to server", slog.Any("error", err))
			time.Sleep(1 * time.Second)
			continue
		}
		c.log.Debug("connected to server", slog.String("addr", c.network+"://"+c.address))

		// setup handlers
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
		if err = jConn.HandleNotification("update2", c.updates2Dispatcher()); err != nil {
			c.log.Warn("fail to setup update2 handler", slog.Any("error", err))
			_ = jConn.Close()
			continue
		}
		if err = jConn.HandleNotification("update", c.updatesDispatcher()); err != nil {
			c.log.Warn("fail to setup update handler", slog.Any("error", err))
			_ = jConn.Close()
			continue
		}

		err = func() error {
			c.monMu.RLock()
			defer c.monMu.RUnlock()
			c.jConn = jConn

			_ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
			defer cancel()

			dbs, err := c.ListDbs(_ctx)
			if err != nil {
				c.log.Warn("fail to list dbs", slog.Any("error", err))
				_ = jConn.Close()
				return err
			}
			c.log.Debug("dbs listed")

			for _, db := range dbs {
				sch, err := c.GetSchema(_ctx, db)
				if err != nil {
					c.log.Warn("fail to get db schema", slog.String("dbName", db), slog.Any("error", err))
					_ = jConn.Close()
					return err
				}
				c.schemas[db] = sch
			}

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

func (c *Client) updates3Dispatcher() func(string, string, monitor.RawTableSetUpdate2) {
	return func(monName string, txnId string, _upd monitor.RawTableSetUpdate2) {
		c.log.Debug("updates3 dispatcher")
		c.monMu.RLock()
		item, ok := c.monitors[monName]
		c.monMu.RUnlock()
		if !ok {
			c.log.Warn("updates3 dispatcher", slog.String("monitor not found", monName))
			return
		}
		item.lastTxnId = txnId

		c.schemasMu.RLock()
		dSch, ok := c.schemas[item.db]
		c.schemasMu.RUnlock()
		if !ok {
			c.log.Warn("updates3 dispatcher: db schema not found", slog.String("name", item.db))
			return
		}

		upd, err := monitor.TableSetUpdateFromRaw2(dSch, _upd)
		if err != nil {
			c.log.Warn("updates3 dispatcher", slog.String("update error", err.Error()))
			return
		}

		select {
		case item.updChan2 <- upd:
		default:
		}
	}
}

func (c *Client) updates2Dispatcher() func(string, monitor.RawTableSetUpdate2) {
	return func(monName string, _upd monitor.RawTableSetUpdate2) {
		c.log.Debug("updates2 dispatcher")
		c.monMu.RLock()
		item, ok := c.monitors[monName]
		c.monMu.RUnlock()
		if !ok {
			c.log.Warn("updates2 dispatcher", slog.String("monitor not found", monName))
			return
		}

		c.schemasMu.RLock()
		dSch, ok := c.schemas[item.db]
		c.schemasMu.RUnlock()
		if !ok {
			c.log.Warn("updates2 dispatcher: db schema not found", slog.String("name", item.db))
			return
		}

		upd, err := monitor.TableSetUpdateFromRaw2(dSch, _upd)
		if err != nil {
			c.log.Warn("updates2 dispatcher", slog.String("update error", err.Error()))
			return
		}

		select {
		case item.updChan2 <- upd:
		default:
		}
	}
}

func (c *Client) updatesDispatcher() func(string, monitor.RawTableSetUpdate) {
	return func(monName string, _upd monitor.RawTableSetUpdate) {
		c.log.Debug("updates dispatcher")
		c.monMu.RLock()
		item, ok := c.monitors[monName]
		c.monMu.RUnlock()
		if !ok {
			c.log.Warn("updates dispatcher", slog.String("monitor not found", monName))
			return
		}
		c.schemasMu.RLock()
		dSch, ok := c.schemas[item.db]
		c.schemasMu.RUnlock()
		if !ok {
			c.log.Warn("updates dispatcher: db schema not found", slog.String("name", item.db))
			return
		}
		upd, err := monitor.TableSetUpdateFromRaw(dSch, _upd)
		if err != nil {
			c.log.Warn("updates dispatcher", slog.String("update error", err.Error()))
			return
		}

		select {
		case item.updChan <- upd:
		default:
		}
	}
}

func (c *Client) Close() error {
	c.closed = true
	err := c.jConn.Close()
	return err
}
