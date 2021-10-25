package controllers

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
)

// findDockercfgSecret checks all the secrets in the namespace to see if the token secret has any existing dockercfg secrets that reference it
func findDockercfgSecrets(client kubernetes.Interface, tokenSecret *corev1.Secret) ([]*corev1.Secret, error) {
	dockercfgSecrets := []*corev1.Secret{}

	// Field constants were removed in the v1.21 API with no immediate replacement.
	// Moving to the external API was marked a TODO, but never implemented.
	// See https://github.com/kubernetes/kubernetes/pull/90105
	options := metav1.ListOptions{FieldSelector: fields.OneTermEqualSelector("type", string(v1.SecretTypeDockercfg)).String()}
	potentialSecrets, err := client.CoreV1().Secrets(tokenSecret.Namespace).List(context.TODO(), options)
	if err != nil {
		return nil, err
	}

	for i, currSecret := range potentialSecrets.Items {
		if currSecret.Annotations[ServiceAccountTokenSecretNameKey] == tokenSecret.Name {
			dockercfgSecrets = append(dockercfgSecrets, &potentialSecrets.Items[i])
		}
	}

	return dockercfgSecrets, nil
}
