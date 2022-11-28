package banned_nodes

import (
	"net/url"
	"time"
)

type BannedNode struct {
	URL        *url.URL
	Timestamp  time.Time
	Expiration time.Time
	Message    string
}
