package webhook

import (
	"openkube/webhook/unitedset/mutating"
	"openkube/webhook/unitedset/validating"
)

func init() {
	addHandlers(mutating.HandlerMap)
	addHandlers(validating.HandlerMap)
}