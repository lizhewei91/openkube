package validating

import (
	"context"
	"fmt"
	"net/http"

	appsv1beta1 "openkube/api/v1beta1"

	"sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

type SidecarSetValidatingHandler struct {
	Decoder *admission.Decoder
}

// podValidator admits a pod iff a specific annotation exists.
func (v *SidecarSetValidatingHandler) Handle(ctx context.Context, req admission.Request) admission.Response {
	log.Log.Info("before unitedSet validating")
	defer log.Log.Info("after unitedSet validating")
	obj := &appsv1beta1.UnitedSet{}

	err := v.Decoder.Decode(req, obj)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	key := "openkube.com/unitedSet-hash"
	_, found := obj.Annotations[key]
	if !found {
		return admission.Denied(fmt.Sprintf("unitedset missing annotation %s", key))
	}

	return admission.Allowed("")
}

var _ admission.DecoderInjector =&SidecarSetValidatingHandler{}

// InjectDecoder injects the decoder.
func (v *SidecarSetValidatingHandler) InjectDecoder(d *admission.Decoder) error {
	v.Decoder = d
	return nil
}
