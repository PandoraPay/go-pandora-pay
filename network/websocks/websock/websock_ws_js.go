//go:build js
// +build js

package websock

import (
	"fmt"
	"syscall/js"
)

type WebSocket struct {
	v js.Value
}

func createWebsocket(url string) (*WebSocket, error) {

	if t := wsJs.Type(); t != js.TypeFunction {
		return nil, fmt.Errorf("constructor is not js.TypeFunction (was %s)", t)
	}

	v := wsJs.New(url)

	if t := v.Type(); t != js.TypeObject {
		return nil, fmt.Errorf("WebSocket type is not js.TypeObject (was %s)", t)
	}

	ws := &WebSocket{
		v,
	}

	ws.setBinaryType("arraybuffer")

	return ws, nil
}

func (ws *WebSocket) onOpen(cb func()) func() {
	return ws.addEventListener("open", func(e js.Value) {
		cb()
	})
}

func (ws *WebSocket) onError(cb func()) func() {
	return ws.addEventListener("error", func(e js.Value) {
		cb()
	})
}

func (ws *WebSocket) onClose(cb func(*CloseEvent)) func() {
	return ws.addEventListener("close", func(e js.Value) {
		cb(&CloseEvent{
			Code:     uint16(e.Get("code").Int()),
			Reason:   e.Get("reason").String(),
			WasClean: e.Get("wasClean").Bool(),
		})
	})
}

func (ws *WebSocket) onMessage(cb func(any)) func() {
	return ws.addEventListener("message", func(e js.Value) {

		var data any

		arrayBuffer := e.Get("data")

		if arrayBuffer.Type() == js.TypeString {
			data = arrayBuffer.String()
		} else {
			data = extractArrayBuffer(arrayBuffer)
		}

		cb(data)

		return
	})
}

func (ws *WebSocket) setBinaryType(typ string) {
	ws.v.Set("binaryType", string(typ))
}

func (ws *WebSocket) addEventListener(eventType string, fn func(e js.Value)) func() {
	f := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		fn(args[0])
		return nil
	})
	ws.v.Call("addEventListener", eventType, f)

	return func() {
		ws.v.Call("removeEventListener", eventType, f)
		f.Release()
	}
}

// ReadyState
func (ws *WebSocket) ReadyState() ReadyState {
	return ReadyState(ws.v.Get("readyState").Int())
}

// BufferedAmount
func (ws *WebSocket) BufferedAmount() int {
	return ws.v.Get("bufferedAmount").Int()
}

// Close
func (ws *WebSocket) Close(code StatusCode, reason string) (err error) {
	defer handleJSError(&err, nil)
	if ws.v.Type() != js.TypeObject {
		ws.v.Call("close", code, reason)
	}
	return
}

// Send
func (ws *WebSocket) SendBytes(v []byte) (err error) {
	defer handleJSError(&err, nil)
	ws.v.Call("send", uint8Array(v))
	return
}

func (ws *WebSocket) SendString(v string) (err error) {
	defer handleJSError(&err, nil)
	ws.v.Call("send", v)
	return
}
