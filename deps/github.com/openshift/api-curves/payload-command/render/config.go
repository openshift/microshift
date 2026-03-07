package render

import (
	"encoding/json"

	configv1 "github.com/openshift/api/config/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
)

var (
	configScheme = runtime.NewScheme()
	configCodecs = serializer.NewCodecFactory(configScheme)
)

func init() {
	utilruntime.Must(configv1.AddToScheme(configScheme))
}

func readFeatureGateV1OrDie(objBytes []byte) *configv1.FeatureGate {
	requiredObj, err := runtime.Decode(configCodecs.UniversalDecoder(configv1.SchemeGroupVersion), objBytes)
	if err != nil {
		panic(err)
	}

	return requiredObj.(*configv1.FeatureGate)
}

func writeFeatureGateV1OrDie(obj *configv1.FeatureGate) string {
	asMap, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
	if err != nil {
		panic(err)
	}
	if _, ok := asMap["apiVersion"]; !ok {
		asMap["apiVersion"] = configv1.GroupVersion.Identifier()
	}
	if _, ok := asMap["kind"]; !ok {
		asMap["kind"] = "FeatureGate"
	}

	ret, err := json.MarshalIndent(asMap, "", "    ")
	if err != nil {
		panic(err)
	}
	return string(ret) + "\n"
}
