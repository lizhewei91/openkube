package webhook

import (
	"openkube/log"

	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

var (
	// HandlerMap contains all admission webhook handlers.
	HandlerMap = map[string]admission.Handler{}
)

func addHandlers(m map[string]admission.Handler) {
	for path, handler := range m {
		if len(path) == 0 {
			log.Error("Skip handler with empty path.")
			continue
		}
		if path[0] != '/' {
			path = "/" + path
		}
		_, found := HandlerMap[path]
		if found {
			log.Infof("conflicting webhook builder path %v in handler map", path)
		}
		HandlerMap[path] = handler
	}
}

func SetupWithManager(mgr manager.Manager) error {
	server := mgr.GetWebhookServer()

	// register admission handlers
	for path, handler := range HandlerMap {
		server.Register(path, &webhook.Admission{Handler: handler})
		log.Infof("Registered webhook handler %s", path)
	}

	return nil
}
