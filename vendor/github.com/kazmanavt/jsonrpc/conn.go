package jsonrpc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// rawConnection is a wrapper around net.Conn that tracks incomplete writes.
type rawConnection struct {
	net.Conn
	failState atomic.Bool
}

func (c *rawConnection) Write(b []byte) (int, error) {
	n, err := c.Conn.Write(b)
	if errors.Is(err, os.ErrDeadlineExceeded) && n < len(b) && n != 0 {
		c.failState.Store(true)
		err = fmt.Errorf("incomplete write: %w", err)
	}
	return n, err
}

// A Connection is a JSON-RPC connection.
type Connection struct {
	defaultTimeout time.Duration
	conn           *rawConnection // underlying connection
	mu             sync.RWMutex   // protects outgoing requests & Close operation
	actionChan     chan *action   // channel for sending requests
	wg             sync.WaitGroup
	log            *slog.Logger
}

// NewConnection creates a new Connection
func NewConnection(network, addr string, _log *slog.Logger) (*Connection, error) {
	if _log == nil {
		_log = slog.Default()
	}

	c, err := net.Dial(network, addr)
	if err != nil {
		return nil, err
	}

	_log.Debug("connected", slog.String("uri", fmt.Sprintf("%s://%s", network, addr)))

	actionChan := make(chan *action, 200)
	conn := &Connection{
		defaultTimeout: 5 * time.Second,
		conn:           &rawConnection{Conn: c},
		actionChan:     actionChan,
		log:            _log,
	}

	notificationChan := make(chan *Response, 200)
	responseChan := make(chan *Response, 200)
	callChan := make(chan *Response, 200)
	conn.wg.Add(2)
	go broker(conn, responseChan, notificationChan, callChan)
	go receiver(conn, responseChan, notificationChan, callChan)

	return conn, nil
}

func broker(c *Connection, responseChan <-chan *Response, notificationChan <-chan *Response, callChan <-chan *Response) {
	defer c.wg.Done()
	var actionChan <-chan *action = c.actionChan
	enc := json.NewEncoder(c.conn)
	nextID := uint64(0)
	pendingRequests := make(map[string]*request)
	pendingCalls := make(map[string]chan<- struct{}, 0)
	defer func() {
		for _, req := range pendingRequests {
			req.res <- &Response{ID: req.ID, Err: c.Error()}
			close(req.res)
		}
		for _, ch := range pendingCalls {
			close(ch)
		}
	}()
	notificationHandlers := make(map[string]NotificationHandler)
	callHandlers := make(map[string]CallHandler)

	for {
		select {
		// send request/notification/call_response or
		// handle internal incoming actions supposed to operate in current goroutine
		case act, ok := <-actionChan:
			if !ok {
				return
			}
			switch act.action {
			case setNotificationHandlerAction:
				// ------------------------------------------
				// handle installation of notification handler
				notificationHandlers[act.method] = act.handler
			case setCallHandlerAction:
				// ------------------------------------------
				// handle installation of call handler
				callHandlers[act.method] = act.callHandler
			case dropPendingRequestAction:
				// ------------------------------------------
				// cancel pending request
				req, ok := pendingRequests[act.hId]
				if !ok {
					continue
				}
				delete(pendingRequests, act.hId)
				req.res <- &Response{ID: req.ID, Err: fmt.Errorf("request cancelled")}
				close(req.res)
			case requestAction:
				// ------------------------------------------
				// handle outgoing request
				var failure bool
				nextID, failure = doSendRequest(c, nextID, act, enc, pendingRequests)
				if failure {
					return
				}
			case notificationAction:
				// ------------------------------------------
				// handle outgoing notification
				if doSendNotification(c, act, enc) {
					return
				}
			case responseAction:
				// ------------------------------------------
				// handle outgoing call response
				if doSendCallResponse(c, act, enc) {
					return
				}
			}

		// receive request result
		case res, ok := <-responseChan:
			if !ok {
				return
			}
			c.log.Debug("received response",
				slog.String("id", *res.ID),
				slog.String("response", string(res.Res)),
				slog.String("error", res.Error().Error()))
			req, ok := pendingRequests[*res.ID]
			if !ok {
				c.log.Debug("unknown response", slog.String("id", *res.ID))
				continue
			}
			delete(pendingRequests, *res.ID)
			req.res <- res
			close(req.res)

		// receive server notification
		case note, ok := <-notificationChan:
			if !ok {
				return
			}
			c.log.Debug("received notification",
				slog.String("method", note.Method),
				slog.String("params", string(note.Params)))
			handler, ok := notificationHandlers[note.Method]
			if !ok {
				c.log.Debug("unknown notification", slog.String("method", note.Method))
				continue
			}
			handler(note.Params)
		case call, ok := <-callChan:
			if !ok {
				return
			}
			c.log.Debug("received call",
				slog.String("id", *call.ID),
				slog.String("method", call.Method),
				slog.String("params", string(call.Params)))
			handler, ok := callHandlers[call.Method]
			if !ok {
				c.log.Debug("unknown call", slog.String("method", call.Method))
				continue
			}

			callRespChan := make(chan *Response)
			stopChan := make(chan struct{})
			pendingCalls[call.Method] = stopChan
			go func(method string) {
				select {
				case <-stopChan:
				case resp := <-callRespChan:
					c.actionChan <- &action{
						action:   responseAction,
						callResp: resp,
					}
					delete(pendingCalls, method)
				}

			}(call.Method)
			handler(call, callRespChan)
		}
	}
}

// doSendRequest sends a request to the server and returns the next request ID to use.
// In case connection became to failed state, it returns second value set to TRUE indicating
// that the caller should stop event loop.
func doSendRequest(c *Connection, nextID uint64, act *action, enc *json.Encoder, pendingRequests map[string]*request) (nextID2 uint64, failed bool) {
	req := newRequest(nextID, act)
	deadline, ok := act.ctx.Deadline()
	if !ok {
		deadline = time.Time{} // no deadline -- use zero time to wait forever
	}
	if err := c.conn.SetWriteDeadline(deadline); err != nil {
		c.log.Warn("fail to set write deadline", slog.String("error", err.Error()))
	}
	nextID++
	fin := make(chan struct{})
	go func() {
		select {
		case <-act.ctx.Done():
			if act.ctx.Err() == context.Canceled {
				_ = c.conn.SetWriteDeadline(time.Now())
				c.log.Debug("request cancelled",
					slog.String("method", act.method),
					slog.String("id", *req.ID))
			}
		case <-fin:
		}
	}()
	err := enc.Encode(req)
	close(fin)
	if err != nil {
		act.idChan <- *req.ID
		close(act.idChan)
		act.respChan <- &Response{ID: req.ID, Err: err}
		close(act.respChan)
		if c.conn.failState.Load() {
			c.log.Debug("connection moved to failed state")
			return 0, true
		}
		return nextID, false
	}

	if c.log.Enabled(context.Background(), slog.LevelDebug) {
		req, _ := json.Marshal(req)
		c.log.Debug("sent request", slog.String("request", string(req)))
	}
	pendingRequests[*req.ID] = req
	act.idChan <- *req.ID
	close(act.idChan)
	return nextID, false
}

// doSendNotification sends a notification to the server.
// In case connection became to failed state, it returns second value set to TRUE indicating
// that the caller should stop event loop.
func doSendNotification(c *Connection, act *action, enc *json.Encoder) bool {
	req := newRequest(0, act)
	deadline, ok := act.ctx.Deadline()
	if !ok {
		deadline = time.Time{} // no deadline -- use zero time to wait forever
	}
	if err := c.conn.SetWriteDeadline(deadline); err != nil {
		c.log.Warn("fail to set write deadline", slog.String("error", err.Error()))
	}
	fin := make(chan struct{})
	go func() {
		select {
		case <-act.ctx.Done():
			if act.ctx.Err() == context.Canceled {
				_ = c.conn.SetWriteDeadline(time.Now())
				c.log.Debug("request cancelled",
					slog.String("method", act.method),
					slog.String("id", *req.ID))
			}
		case <-fin:
		}
	}()
	err := enc.Encode(req)
	close(fin)
	if err != nil {
		act.respChan <- &Response{Err: err}
		close(act.respChan)
		if c.conn.failState.Load() {
			c.log.Debug("connection moved to failed state")
			return true
		}
		return false
	}
	if c.log.Enabled(context.Background(), slog.LevelDebug) {
		req, _ := json.Marshal(req)
		c.log.Debug("sent notification", slog.String("request", string(req)))
	}
	act.respChan <- &Response{}
	close(act.respChan)
	return false
}

func doSendCallResponse(c *Connection, act *action, enc *json.Encoder) bool {
	deadline := time.Now().Add(c.defaultTimeout)
	if err := c.conn.SetWriteDeadline(deadline); err != nil {
		c.log.Warn("fail to set write deadline", slog.String("error", err.Error()))
	}
	err := enc.Encode(act.callResp)
	if err != nil {
		if c.conn.failState.Load() {
			c.log.Debug("connection moved to failed state")
			return true
		}
		c.log.Debug("failed to send call response", slog.String("error", err.Error()))
		return false
	}
	if c.log.Enabled(context.Background(), slog.LevelDebug) {
		resp, _ := json.Marshal(act.callResp)
		c.log.Debug("sent call response", slog.String("response", string(resp)))
	}
	return false
}

// receiver receives JSON-RPC responses and server notifications from the connection.
// And calls from server if any.
func receiver(c *Connection, responseChan chan<- *Response, notificationChan chan<- *Response, callChan chan<- *Response) {
	defer c.wg.Done()
	defer close(notificationChan)
	defer close(responseChan)

	dec := json.NewDecoder(c.conn)

	for {
		var resp Response
		if err := dec.Decode(&resp); err != nil {
			if errors.Is(err, io.EOF) || (errors.Is(err, os.ErrDeadlineExceeded) && c.conn.failState.Load()) ||
				strings.Contains(err.Error(), "use of closed network connection") {
				c.log.Debug("broken connection", slog.String("error", err.Error()))
				c.conn.failState.Store(true)
				return
			}
			c.log.Debug("fail to decode response", slog.String("error", err.Error()))
			continue
		}
		if resp.IsNotification() {
			// notification
			notificationChan <- &resp
			continue
		}
		if resp.IsCall() {
			// call
			callChan <- &resp
			continue
		}
		responseChan <- &resp
	}
}

func (c *Connection) Error() error {
	var err error = nil
	switch {
	//case c == nil:
	//	err = errors.New("nil connection")
	//case c.conn == nil:
	//	err = errors.New("closed connection")
	case c.conn.failState.Load():
		err = errors.New("failed connection")
	}
	return err
}

// Close closes the connection.
func (c *Connection) Close() error {
	//if c == nil {
	//	c.log.Errorw("close on nil connection")
	//	return errors.New("close on nil connection")
	//}
	//if c.conn == nil {
	//	c.log.Debug("close on closed connection")
	//	return nil
	//}
	c.mu.Lock()
	defer c.mu.Unlock()
	err := c.conn.Close()
	c.wg.Wait()
	close(c.actionChan)
	c.log.Debug("connection closed")
	return err
}

// validateContext check if context not nil and set default timeout if needed.
func (c *Connection) validateContext(ctx context.Context) context.Context {
	if ctx == nil {
		ctx, _ = context.WithTimeout(context.Background(), c.defaultTimeout)
	}
	return ctx
}

func (c *Connection) Notify(ctx context.Context, method string, params ...any) error {
	//if c == nil || c.conn == nil {
	//	return errors.New("closed or nil connection")
	//}
	c.mu.RLock()
	defer c.mu.RUnlock()

	if err := c.Error(); err != nil {
		return err
	}

	respChan := make(chan *Response, 1)
	c.actionChan <- &action{
		action:   notificationAction,
		method:   method,
		params:   params,
		ctx:      c.validateContext(ctx),
		respChan: respChan,
	}
	resp := <-respChan
	if resp.Err != nil {
		return resp.Error()
	}
	return nil
}

// Send sends a single JSON-RPC request asynchronously.
func (c *Connection) Send(ctx context.Context, method string, params ...any) (<-chan *Response, string, error) {
	//if c == nil || c.conn == nil {
	//	return nil, "", errors.New("closed or nil connection")
	//}
	c.mu.RLock()
	defer c.mu.RUnlock()

	if err := c.Error(); err != nil {
		return nil, "", err
	}

	respChan := make(chan *Response, 1)
	idChan := make(chan string)
	c.actionChan <- &action{
		action:   requestAction,
		method:   method,
		params:   params,
		ctx:      c.validateContext(ctx),
		idChan:   idChan,
		respChan: respChan,
	}
	return respChan, <-idChan, nil
}

func (c *Connection) DropPending(id string) {
	//if c == nil || c.conn == nil {
	//	return
	//}
	c.actionChan <- &action{
		action: dropPendingRequestAction,
		hId:    id,
	}
}

// Call sends a single JSON-RPC request synchronously.
func (c *Connection) Call(ctx context.Context, method string, params ...any) (json.RawMessage, error) {
	//if c == nil || c.conn == nil {
	//	return nil, errors.New("closed or nil connection")
	//}
	ctx = c.validateContext(ctx)
	respChan, id, err := c.Send(ctx, method, params...)
	if err != nil {
		return nil, err
	}
	// wait for response or context cancellation
	select {
	case res := <-respChan:
		return res.Res, res.Error()
	case <-ctx.Done():
		c.DropPending(id)
		return nil, ctx.Err()
	}
}

// Handle sets notification handler for incoming JSON-RPC notification.
func (c *Connection) Handle(method string, handler NotificationHandler) error {
	//if c == nil || c.conn == nil {
	//	return errors.New("closed or nil connection")
	//}
	c.mu.RLock()
	defer c.mu.RUnlock()
	if err := c.Error(); err != nil {
		return err
	}
	c.actionChan <- &action{
		action:  setNotificationHandlerAction,
		method:  method,
		handler: handler,
	}
	return nil
}

// HandleCall sets call handler for incoming JSON-RPC call.
func (c *Connection) HandleCall(method string, handler CallHandler) error {
	//if c == nil || c.conn == nil {
	//	return errors.New("closed or nil connection")
	//}
	c.mu.RLock()
	defer c.mu.RUnlock()
	if err := c.Error(); err != nil {
		return err
	}
	c.actionChan <- &action{
		action:      setCallHandlerAction,
		method:      method,
		callHandler: handler,
	}
	return nil
}
