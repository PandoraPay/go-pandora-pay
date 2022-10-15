package globals

import (
	"market/helpers/generics"
	"pandora-pay/helpers/events"
)

// arguments
var (
	Arguments   map[string]interface{}
	MainEvents  = events.NewEvents[any]()
	MainStarted = generics.Value[bool]{}
)
