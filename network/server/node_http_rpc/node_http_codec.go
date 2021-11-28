package node_http_rpc

import (
	"github.com/gorilla/rpc"
	"github.com/gorilla/rpc/json"
	"net/http"
)

// UpCodec creates a CodecRequest to process each request.
type UpCodec struct {
}

// NewUpCodec returns a new UpCodec.
func NewUpCodec() *UpCodec {
	return &UpCodec{}
}

// NewRequest returns a new CodecRequest of type UpCodecRequest.
func (c *UpCodec) NewRequest(r *http.Request) rpc.CodecRequest {
	outerCR := &UpCodecRequest{}   // Our custom CR
	jsonC := json.NewCodec()       // json Codec to create json CR
	innerCR := jsonC.NewRequest(r) // create the json CR, sort of.

	// NOTE - innerCR is of the interface type rpc.CodecRequest.
	// Because innerCR is of the rpc.CR interface type, we need a
	// type assertion in order to assign it to our struct field's type.
	// We defined the source of the interface implementation here, so
	// we can be confident that innerCR will be of the correct underlying type
	outerCR.CodecRequest = innerCR.(*json.CodecRequest)
	return outerCR
}

// UpCodecRequest decodes and encodes a single request. UpCodecRequest
// implements gorilla/rpc.CodecRequest interface primarily by embedding
// the CodecRequest from gorilla/rpc/json. By selectively adding
// CodecRequest methods to UpCodecRequest, we can modify that behaviour
// while maintaining all the other remaining CodecRequest methods from
// gorilla's rpc/json implementation
type UpCodecRequest struct {
	*json.CodecRequest
}

// Method returns the decoded method as a string of the form "Service.Method"
// after checking for, and correcting a lowercase method name
// By being of lower depth in the struct , Method will replace the implementation
// of Method() on the embedded CodecRequest. Because the request data is part
// of the embedded json.CodecRequest, and unexported, we have to get the
// requested method name via the embedded CR's own method Method().
// Essentially, this just intercepts the return value from the embedded
// gorilla/rpc/json.CodecRequest.Method(), checks/modifies it, and passes it
// on to the calling rpc server.
func (c *UpCodecRequest) Method() (string, error) {
	m, err := c.CodecRequest.Method()
	if len(m) > 1 && err == nil {

		final := make([]byte, len(m))
		c := 0
		for i := 0; i < len(m); i++ {
			if m[i] == '/' || m[i] == '-' {
				final[c] = m[i+1] - 32
				c += 1
				i += 1
				continue
			} else if i == 0 {
				final[c] = m[0] - 32
				c += 1
			} else {
				final[c] = m[i]
				c += 1
			}
		}

		upMethod := "api." + string(final[:c])
		return upMethod, err
	}
	return m, err
}
