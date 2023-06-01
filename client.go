package ovsdb

import (
	"encoding/json"
	"fmt"
	"github.com/kazmanavt/jsonrpc"
	"go.uber.org/zap"
	"sync"
)

type ClientConf struct {
	Log  *zap.SugaredLogger
	Conn *jsonrpc.Connection
	Net  string
	Addr string
}
type Client struct {
	*jsonrpc.Connection
	log             *zap.SugaredLogger
	updatesHandlers map[string]chan<- json.RawMessage
	uhMu            sync.Mutex
}

func NewClient(conf ClientConf) (c *Client, err error) {
	if conf.Log == nil {
		conf.Log = zap.NewNop().Sugar()
	}
	if conf.Conn == nil {
		_jrpcLog := conf.Log.Named("JSON-RPC")
		conf.Log.Debugw("creating new connection", zap.String("net", conf.Net), zap.String("addr", conf.Addr))
		conf.Conn, err = jsonrpc.NewConnection(conf.Net, conf.Addr, _jrpcLog)
		if err != nil {
			return nil, err
		}
	}
	c = &Client{
		Connection:      conf.Conn,
		log:             conf.Log,
		updatesHandlers: make(map[string]chan<- json.RawMessage, 10),
	}
	defer func() {
		if err != nil {
			err2 := c.Close()
			err = fmt.Errorf("%w, and close error: %v", err, err2)
		}
	}()

	if err = c.Connection.HandleCall("echo", c.echoHandler); err != nil {
		return nil, err
	}
	if err = c.Connection.Handle("update", c.updatesDispatcher); err != nil {
		return nil, err
	}
	if err = c.Connection.Handle("update2", c.updatesDispatcher); err != nil {
		return nil, err
	}
	if err = c.Connection.Handle("update3", c.updatesDispatcher); err != nil {
		return nil, err
	}

	return c, nil
}

func (c *Client) echoHandler(req *jsonrpc.Response, respChan chan<- *jsonrpc.Response) {
	c.log.Debugw("echo handler", zap.Any("server req.params", req.Params))
	respChan <- &jsonrpc.Response{
		ID:  req.ID,
		Res: req.Params,
	}
}

func (c *Client) updatesDispatcher(msg []byte) {
	c.log.Debugw("updates dispatcher")
	update := []any{new(string)}
	err := json.Unmarshal(msg, &update)
	if err != nil {
		c.log.Errorw("fail to unmarshal update (phase 0)", zap.Error(err))
		return
	}
	id, ok := update[0].(*string)
	if !ok {
		c.log.Errorw("malformed update notification: fail to get id")
		return
	}
	c.uhMu.Lock()
	handler, ok := c.updatesHandlers[*id]
	defer c.uhMu.Unlock()
	if !ok {
		c.log.Warnw("no handler for update", zap.String("id", *id))
		return
	}
	select {
	case handler <- msg:
	default:
	}
}

func (c *Client) Close() error {
	err := c.Connection.Close()
	c.uhMu.Lock()
	defer c.uhMu.Unlock()
	for _, handler := range c.updatesHandlers {
		close(handler)
	}
	return err
}
