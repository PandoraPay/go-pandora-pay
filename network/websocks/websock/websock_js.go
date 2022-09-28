//go:build js
// +build js

package websock

import (
	"context"
	"errors"
	"fmt"
	"pandora-pay/config"
	"pandora-pay/helpers/generics"
	"sync"
	"time"
)

type Conn struct {
	ws               *WebSocket
	limit            *generics.Value[int64]
	opened           chan struct{}
	closed           chan struct{}
	closedErrorOnce  *sync.Once
	closedErr        *generics.Value[error]
	readCtx          context.Context
	releaseOnClose   func()
	releaseOnMessage func()
	releaseOnOpen    func()
	releaseOnError   func()
	readSignal       chan struct{}
	readBufMu        sync.Mutex
	readBuf          []any
}

func Dial(url string) (c *Conn, err error) {

	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("constructor error %v", r)
		}
	}()

	ws, err := createWebsocket(url)
	if err != nil {
		return nil, err
	}

	c = &Conn{
		ws,
		&generics.Value[int64]{},
		make(chan struct{}),
		make(chan struct{}),
		&sync.Once{},
		&generics.Value[error]{},
		nil,
		func() {},
		func() {},
		func() {},
		func() {},
		make(chan struct{}),
		sync.Mutex{},
		make([]any, 0),
	}

	c.limit.Store(32 * 1024)

	c.releaseOnClose = c.ws.onClose(func(e *CloseEvent) {

		c.closedEvent(&CloseError{
			Code:   StatusCode(e.Code),
			Reason: e.Reason,
		}, e.WasClean)

		c.releaseOnClose()
		c.releaseOnMessage()
		c.releaseOnOpen()
		c.releaseOnError()
	})

	c.releaseOnMessage = c.ws.onMessage(func(data any) {
		c.readBufMu.Lock()
		defer c.readBufMu.Unlock()

		c.readBuf = append(c.readBuf, data)

		// Lets the read goroutine know there is definitely something in readBuf.
		select {
		case c.readSignal <- struct{}{}:
		default:
		}
	})

	c.releaseOnOpen = c.ws.onOpen(func() {
		close(c.opened)
	})

	c.releaseOnError = c.ws.onError(func() {
		fmt.Println("Web Socket error")
	})

	ctx, cancel := context.WithTimeout(context.Background(), config.WEBSOCKETS_TIMEOUT)
	defer cancel()

	select {
	case <-ctx.Done():
		c.ws.Close(StatusPolicyViolation, "dial timed out")
		return nil, ctx.Err()
	case <-c.opened:
		return c, nil
	}

	return
}

func (c *Conn) SetPongHandler(cb func(string) error) error {
	go func() {
		for {
			select {
			case <-c.closed:
				break
			default:
				cb("PING")
				time.Sleep(config.WEBSOCKETS_PING_INTERVAL)
			}
		}
	}()
	return nil
}

func (c *Conn) SetReadLimit(limit int64) error {
	c.limit.Store(limit)
	return nil
}

func (c *Conn) SetWriteDeadline(limit time.Time) error {
	return nil
}

func (c *Conn) SetReadDeadline(limit time.Time) error {
	c.readCtx, _ = context.WithDeadline(context.Background(), limit)
	return nil
}

func (c *Conn) ReadMessage() (int, []byte, error) {
	select {
	case <-c.readCtx.Done():
		c.ws.Close(StatusPolicyViolation, "read timed out")
		return 0, nil, c.readCtx.Err()
	case <-c.readSignal:
	case <-c.closed:
		return 0, nil, c.closedErr.Load()
	}

	c.readBufMu.Lock()
	defer c.readBufMu.Unlock()

	data := c.readBuf[0]
	// We copy the messages forward and decrease the size
	// of the slice to avoid reallocating.
	copy(c.readBuf, c.readBuf[1:])
	c.readBuf = c.readBuf[:len(c.readBuf)-1]

	if len(c.readBuf) > 0 {
		// Next time we read, we'll grab the message.
		select {
		case c.readSignal <- struct{}{}:
		default:
		}
	}

	switch p := data.(type) {
	case string:
		return TextMessage, []byte(p), nil
	case []byte:
		return BinaryMessage, p, nil
	default:
		panic("websocket: unexpected data type")
	}

	return 0, nil, nil
}

func (c *Conn) WriteMessage(msg int, data []byte) error {
	c.ws.SendBytes(data)
	return nil
}

func (c *Conn) Close() error {
	c.ws.Close(0, "close")
	return nil
}

func (c *Conn) closedEvent(e *CloseError, wasClean bool) {
	c.closedErrorOnce.Do(func() {
		close(c.closed)
		if e.Reason != "" {
			c.closedErr.Store(errors.New(e.Reason))
		}
	})
}

func (c *Conn) isClosed() bool {
	select {
	case <-c.closed:
		return true
	default:
		return false
	}
}
