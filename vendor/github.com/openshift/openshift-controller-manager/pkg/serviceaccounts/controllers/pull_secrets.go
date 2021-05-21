package controllers

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	api "k8s.io/kubernetes/pkg/apis/core"
)

// findDockercfgSecret checks all the secrets in the namespace to see if the token secret has any existing dockercfg secrets that reference it
func findDockercfgSecrets(client kubernetes.Interface, tokenSecret *corev1.Secret) ([]*corev1.Secret, error) {
	dockercfgSecrets := []*corev1.Secret{}

	options := metav1.ListOptions{FieldSelector: fields.OneTermEqualSelector(api.SecretTypeField, string(v1.SecretTypeDockercfg)).String()}
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
