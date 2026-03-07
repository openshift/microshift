package resourceread

import (
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
)

var (
	apiExtensionsScheme = runtime.NewScheme()
	apiExtensionsCodecs = serializer.NewCodecFactory(apiExtensionsScheme)
)

func init() {
	utilruntime.Must(apiextensionsv1.AddToScheme(apiExtensionsScheme))
}

func ReadCustomResourceDefinitionV1(objBytes []byte) (*apiextensionsv1.CustomResourceDefinition, error) {
	requiredObj, err := runtime.Decode(apiExtensionsCodecs.UniversalDecoder(apiextensionsv1.SchemeGroupVersion), objBytes)
	if err != nil {
		return nil, err
	}
	return requiredObj.(*apiextensionsv1.CustomResourceDefinition), nil
}

func ReadCustomResourceDefinitionV1OrDie(objBytes []byte) *apiextensionsv1.CustomResourceDefinition {
	requiredObj, err := runtime.Decode(apiExtensionsCodecs.UniversalDecoder(apiextensionsv1.SchemeGroupVersion), objBytes)
	if err != nil {
		panic(err)
	}
	return requiredObj.(*apiextensionsv1.CustomResourceDefinition)
}

func WriteCustomResourceDefinitionV1OrDie(obj *apiextensionsv1.CustomResourceDefinition) string {
	return runtime.EncodeOrDie(apiExtensionsCodecs.LegacyCodec(apiextensionsv1.SchemeGroupVersion), obj)
}
