package globals

import (
	"pandora-pay/helpers/events"
	"pandora-pay/helpers/generics"
)

// arguments
var (
	MainEvents  = events.NewEvents[any]()
	MainStarted = generics.Value[bool]{}
)

func init() {
	MainStarted.Store(false)
}
