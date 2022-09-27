//go:build js
// +build js

package websock

import (
	"syscall/js"
)

var wsJs js.Value

// ReadyState https://developer.mozilla.org/en-US/docs/Web/API/WebSocket/readyState
type ReadyState uint16

const (
	ReadyStateConnecting = iota
	ReadyStateOpen
	ReadyStateClosing
	ReadyStateClosed
)

// StatusCode represents a WebSocket status code.
// https://tools.ietf.org/html/rfc6455#section-7.4
type StatusCode int

// https://www.iana.org/assignments/websocket/websocket.xhtml#close-code-number
//
// These are only the status codes defined by the protocol.
//
// You can define custom codes in the 3000-4999 range.
// The 3000-3999 range is reserved for use by libraries, frameworks and applications.
// The 4000-4999 range is reserved for private use.
const (
	StatusNormalClosure   StatusCode = 1000
	StatusGoingAway       StatusCode = 1001
	StatusProtocolError   StatusCode = 1002
	StatusUnsupportedData StatusCode = 1003

	// 1004 is reserved and so unexported.
	statusReserved StatusCode = 1004

	// StatusNoStatusRcvd cannot be sent in a close message.
	// It is reserved for when a close message is received without
	// a status code.
	StatusNoStatusRcvd StatusCode = 1005

	// StatusAbnormalClosure is exported for use only with Wasm.
	// In non Wasm Go, the returned error will indicate whether the
	// connection was closed abnormally.
	StatusAbnormalClosure StatusCode = 1006

	StatusInvalidFramePayloadData StatusCode = 1007
	StatusPolicyViolation         StatusCode = 1008
	StatusMessageTooBig           StatusCode = 1009
	StatusMandatoryExtension      StatusCode = 1010
	StatusInternalError           StatusCode = 1011
	StatusServiceRestart          StatusCode = 1012
	StatusTryAgainLater           StatusCode = 1013
	StatusBadGateway              StatusCode = 1014

	// StatusTLSHandshake is only exported for use with Wasm.
	// In non Wasm Go, the returned error will indicate whether there was
	// a TLS handshake failure.
	StatusTLSHandshake StatusCode = 1015
)

type CloseError struct {
	Code   StatusCode
	Reason string
}

type CloseEvent struct {
	Code     uint16
	Reason   string
	WasClean bool
}

func handleJSError(err *error, onErr func()) {
	r := recover()

	if jsErr, ok := r.(js.Error); ok {
		*err = jsErr

		if onErr != nil {
			onErr()
		}
		return
	}

	if r != nil {
		panic(r)
	}
}

func extractArrayBuffer(arrayBuffer js.Value) []byte {
	uint8Array := js.Global().Get("Uint8Array").New(arrayBuffer)
	dst := make([]byte, uint8Array.Length())
	js.CopyBytesToGo(dst, uint8Array)
	return dst
}

func uint8Array(src []byte) js.Value {
	uint8Array := js.Global().Get("Uint8Array").New(len(src))
	js.CopyBytesToJS(uint8Array, src)
	return uint8Array
}

func init() {
	wsJs = js.Global().Get("WebSocket")
}
