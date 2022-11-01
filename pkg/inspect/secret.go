package inspect

import (
	"os"
	"path"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/cli-runtime/pkg/resource"
)

func inspectSecretInfo(info *resource.Info, o *InspectOptions) error {
	obj := info.Object

	if unstructureObj, ok := obj.(*unstructured.Unstructured); ok {
		structuredSecret := &corev1.Secret{}
		err := runtime.DefaultUnstructuredConverter.FromUnstructured(unstructureObj.Object, structuredSecret)
		if err != nil {
			return err
		}
		obj = structuredSecret
	}
	if unstructureObjList, ok := obj.(*unstructured.UnstructuredList); ok {
		structuredSecretList := &corev1.SecretList{}
		err := runtime.DefaultUnstructuredConverter.FromUnstructured(unstructureObjList.Object, structuredSecretList)
		if err != nil {
			return err
		}
		for _, unstructureObj := range unstructureObjList.Items {
			structuredSecret := &corev1.Secret{}
			err := runtime.DefaultUnstructuredConverter.FromUnstructured(unstructureObj.Object, structuredSecret)
			if err != nil {
				return err
			}
			structuredSecretList.Items = append(structuredSecretList.Items, *structuredSecret)
		}

		obj = structuredSecretList
	}

	switch castObj := obj.(type) {
	case *corev1.Secret:
		elideSecret(castObj)

	case *corev1.SecretList:
		for i := range castObj.Items {
			elideSecret(&castObj.Items[i])
		}

	case *unstructured.UnstructuredList:

	}

	// save the current object to disk
	dirPath := dirPathForInfo(o.destDir, info)
	filename := filenameForInfo(info)
	// ensure destination path exists
	if err := os.MkdirAll(dirPath, os.ModePerm); err != nil {
		return err
	}
	return o.fileWriter.WriteFromResource(path.Join(dirPath, filename), obj)
}

var publicSecretKeys = sets.NewString(
	// we know that tls.crt contains certificate (public) data, not private data.  This allows inspection of signing names for signers.
	"tls.crt",
	// we know that ca.crt contains certificate (public) data, not private data.  This allows inspection of sa token ca.crt trust.
	"ca.crt",
	// we know that service-ca.crt contains certificate (public) data, not private data.  This allows inspection of sa token service-ca.crt trust.
	"service-ca.crt",
)

func elideSecret(secret *corev1.Secret) {
	for k := range secret.Data {
		// some secrets keys are safe to include because know their content.
		if publicSecretKeys.Has(k) {
			continue
		}
		secret.Data[k] = []byte{}
	}

	if _, ok := secret.Annotations["openshift.io/token-secret.value"]; ok {
		secret.Annotations["openshift.io/token-secret.value"] = ""
	}
	if _, ok := secret.Annotations["kubectl.kubernetes.io/last-applied-configuration"]; ok {
		secret.Annotations["kubectl.kubernetes.io/last-applied-configuration"] = ""
	}
}
