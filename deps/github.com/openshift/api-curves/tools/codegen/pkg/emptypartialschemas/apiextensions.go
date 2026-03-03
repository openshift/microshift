package emptypartialschemas

import (
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"sigs.k8s.io/yaml"
)

var (
	apiExtensionsScheme = runtime.NewScheme()
	apiExtensionsCodecs = serializer.NewCodecFactory(apiExtensionsScheme)
)

func init() {
	utilruntime.Must(apiextensionsv1beta1.AddToScheme(apiExtensionsScheme))
	utilruntime.Must(apiextensionsv1.AddToScheme(apiExtensionsScheme))
}

func ReadCustomResourceDefinitionV1OrDie(objBytes []byte) *apiextensionsv1.CustomResourceDefinition {
	requiredObj, err := ReadCustomResourceDefinitionV1(objBytes)
	if err != nil {
		panic(err)
	}
	return requiredObj
}

func ReadCustomResourceDefinitionV1(objBytes []byte) (*apiextensionsv1.CustomResourceDefinition, error) {
	// very funky, but with the normal decode path we're getting defaulting
	uncast := map[string]interface{}{}
	if err := yaml.Unmarshal(objBytes, &uncast); err != nil {
		return nil, err
	}
	ret := &apiextensionsv1.CustomResourceDefinition{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(uncast, ret); err != nil {
		return nil, err
	}

	return ret, nil
}

func WriteSpecOnlyCustomResourceDefinitionV1(obj *apiextensionsv1.CustomResourceDefinition) ([]byte, error) {
	uncastObj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
	if err != nil {
		return nil, err
	}
	uncastObj["kind"] = "CustomResourceDefinition"
	uncastObj["apiVersion"] = apiextensionsv1.SchemeGroupVersion.Identifier()
	delete(uncastObj, "status")
	unstructured.RemoveNestedField(uncastObj, "metadata", "creationTimestamp")
	return yaml.Marshal(uncastObj)
}
