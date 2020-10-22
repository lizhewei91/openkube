package mutating

import(
	"crypto/sha256"
	"encoding/json"
	"fmt"

	appsv1beta1 "openkube/api/v1beta1"

	"k8s.io/apimachinery/pkg/util/rand"

)

func UnitedSetHash(sidecarSet *appsv1beta1.UnitedSet) (string, error) {
	encoded, err := encodeSidecarSet(sidecarSet)
	if err != nil {
		return "", err
	}
	h := rand.SafeEncodeString(hash(encoded))
	return h, nil
}


func encodeSidecarSet(UnitedSet *appsv1beta1.UnitedSet) (string, error) {
	// json.Marshal sorts the keys in a stable order in the encoding
	m := map[string]interface{}{"containers": UnitedSet.Spec.Foo}
	data, err := json.Marshal(m)
	if err != nil {
		return "", err
	}
	return string(data), nil
}


// hash hashes `data` with sha256 and returns the hex string
func hash(data string) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(data)))
}
