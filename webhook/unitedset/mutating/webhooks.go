package mutating

import (
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// +kubebuilder:webhook:path=/mutate-apps-openkube-com-v1beta1-unitedset,mutating=true,failurePolicy=fail,groups=apps.openKube.com,resources=unitedsets,verbs=create;update,versions=v1beta1,name=munitedset.kb.io

var (
	// HandlerMap contains admission webhook handlers
	HandlerMap = map[string]admission.Handler{
		"mutate-apps-openkube-com-v1beta1-unitedset": &UnitedSetMutatingHandler{},
	}
)
