package inspect

import (
	"fmt"
	"os"
	"path"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/cli-runtime/pkg/resource"
)

type secretList struct {
	*corev1.SecretList
}

func (c *secretList) addItem(obj interface{}) error {
	structuredItem, ok := obj.(*corev1.Secret)
	if !ok {
		return fmt.Errorf("unhandledStructuredItemType: %T", obj)
	}
	c.Items = append(c.Items, *structuredItem)
	return nil
}

func inspectSecretInfo(info *resource.Info, o *InspectOptions) error {
	structuredObj, err := toStructuredObject[corev1.Secret, corev1.SecretList](info.Object)
	if err != nil {
		return err
	}

	switch castObj := structuredObj.(type) {
	case *corev1.Secret:
		elideSecret(castObj)

	case *corev1.SecretList:
		for i := range castObj.Items {
			elideSecret(&castObj.Items[i])
		}

	}

	// save the current object to disk
	dirPath := dirPathForInfo(o.DestDir, info)
	filename := filenameForInfo(info)
	// ensure destination path exists
	if err := os.MkdirAll(dirPath, os.ModePerm); err != nil {
		return err
	}
	return o.fileWriter.WriteFromResource(path.Join(dirPath, filename), structuredObj)
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
	for k, v := range secret.Data {
		// some secrets keys are safe to include because know their content.
		if publicSecretKeys.Has(k) {
			continue
		}
		secret.Data[k] = []byte(fmt.Sprintf("%d bytes long", len(v)))
	}

	if _, ok := secret.Annotations["openshift.io/token-secret.value"]; ok {
		secret.Annotations["openshift.io/token-secret.value"] = ""
	}
	if _, ok := secret.Annotations["kubectl.kubernetes.io/last-applied-configuration"]; ok {
		secret.Annotations["kubectl.kubernetes.io/last-applied-configuration"] = ""
	}
}
