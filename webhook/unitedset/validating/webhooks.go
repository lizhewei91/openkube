package validating

import (
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)


// +kubebuilder:webhook:verbs=create;update,path=/validate-apps-openkube-com-v1beta1-unitedset,mutating=false,failurePolicy=fail,groups=apps.openKube.com,resources=unitedsets,versions=v1beta1,name=vunitedset.kb.io

var (
	// HandlerMap contains admission webhook handlers
	HandlerMap = map[string]admission.Handler{
		"validate-apps-openkube-com-v1beta1-unitedset": &SidecarSetValidatingHandler{},
	}
)
