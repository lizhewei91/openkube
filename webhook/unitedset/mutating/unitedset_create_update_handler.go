package mutating

import (
	"context"
	"encoding/json"
	"net/http"

	appsv1beta1 "openkube/api/v1beta1"
	//"openkube/log"

	"k8s.io/api/admission/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

const(
	UnitedSetAnnotation = "openkube.com/unitedSet-hash"
)

type UnitedSetMutatingHandler struct {
	Client client.Client

	Decoder *admission.Decoder
}

var _ admission.Handler = &UnitedSetMutatingHandler{}

func (a *UnitedSetMutatingHandler) Handle(ctx context.Context, req admission.Request) admission.Response {
	log.Log.Info("before unitedSet mutating")
	defer log.Log.Info("after unitedSet mutating")

	obj:=&appsv1beta1.UnitedSet{}

	err := a.Decoder.Decode(req, obj)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	switch req.AdmissionRequest.Operation {
	case v1beta1.Create, v1beta1.Update:
		//appsv1beta1.SetDefaultsSidecarSet(obj)
		if err := setHashUnitedSet(obj); err != nil {
			//log.Errorf("unitedset mutating handle set hash error,err:%v", err)
			return admission.Errored(http.StatusInternalServerError, err)
		}
	}

	marshaledPod, err := json.Marshal(obj)
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}

	return admission.PatchResponseFromRaw(req.Object.Raw, marshaledPod)
}


// set unitedset hash
func setHashUnitedSet(unitedset *appsv1beta1.UnitedSet) error {
	if unitedset.Annotations == nil {
		unitedset.Annotations = make(map[string]string)
	}

	hash, err := UnitedSetHash(unitedset)
	if err != nil {
		return err
	}
	unitedset.Annotations[UnitedSetAnnotation] = hash

	return nil
}

var _ admission.DecoderInjector =&UnitedSetMutatingHandler{}

// InjectDecoder injects the decoder.
func (a *UnitedSetMutatingHandler) InjectDecoder(d *admission.Decoder) error {
	a.Decoder = d
	return nil
}