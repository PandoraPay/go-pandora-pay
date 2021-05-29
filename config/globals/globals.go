package globals

import "pandora-pay/helpers/events"

// arguments
var (
	Arguments   map[string]interface{}
	Data        = map[string]interface{}{}
	MainEvents  = events.NewEvents()
	MainStarted = false
)
