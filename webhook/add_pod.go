package webhook

import (
	"openkube/webhook/pod/mutating"
)

func init() {
	addHandlers(mutating.HandlerMap)
}