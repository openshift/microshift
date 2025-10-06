package components

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/openshift/microshift/pkg/config"
	"github.com/openshift/microshift/pkg/util/cryptomaterial"
)

const (
	// Resource namespace
	caResourceNamespace = "kube-system"
)

var (
	CertificateAuthorityResources = []struct {
		Name string
		Dir  string
	}{
		{Name: "kube-control-plane-signer", Dir: cryptomaterial.KubeControlPlaneSignerCertDir(cryptomaterial.CertsDirectory(config.DataDir))},
		{Name: "kube-apiserver-to-kubelet-signer", Dir: cryptomaterial.KubeAPIServerToKubeletSignerCertDir(cryptomaterial.CertsDirectory(config.DataDir))},
		{Name: "admin-kubeconfig-signer", Dir: cryptomaterial.AdminKubeconfigSignerDir(cryptomaterial.CertsDirectory(config.DataDir))},
		{Name: "kubelet-signer", Dir: cryptomaterial.KubeletCSRSignerSignerCertDir(cryptomaterial.CertsDirectory(config.DataDir))},
		{Name: "kube-csr-signer", Dir: cryptomaterial.CSRSignerCertDir(cryptomaterial.CertsDirectory(config.DataDir))},
		{Name: "aggregator-signer", Dir: cryptomaterial.AggregatorSignerDir(cryptomaterial.CertsDirectory(config.DataDir))},
		{Name: "service-ca", Dir: cryptomaterial.ServiceCADir(cryptomaterial.CertsDirectory(config.DataDir))},
		{Name: "ingress-ca", Dir: cryptomaterial.IngressCADir(cryptomaterial.CertsDirectory(config.DataDir))},
		{Name: "kube-apiserver-external-signer", Dir: cryptomaterial.KubeAPIServerExternalSigner(cryptomaterial.CertsDirectory(config.DataDir))},
		{Name: "kube-apiserver-localhost-signer", Dir: cryptomaterial.KubeAPIServerLocalhostSigner(cryptomaterial.CertsDirectory(config.DataDir))},
		{Name: "kube-apiserver-service-network-signer", Dir: cryptomaterial.KubeAPIServerServiceNetworkSigner(cryptomaterial.CertsDirectory(config.DataDir))},
		{Name: "etcd-signer", Dir: cryptomaterial.EtcdSignerDir(cryptomaterial.CertsDirectory(config.DataDir))},
	}
)

func startCertificateAuthorityController(ctx context.Context, kubeconfigPath string) error {
	client, err := getKubernetesClient(kubeconfigPath)
	if err != nil {
		return fmt.Errorf("failed to get Kubernetes client: %w", err)
	}

	resourceNames := make([]string, len(CertificateAuthorityResources))
	for i, resource := range CertificateAuthorityResources {
		resourceNames[i] = resource.Name
		if err := exposeCertificateAuthority(ctx, client, resource.Dir, resource.Name); err != nil {
			return fmt.Errorf("failed to expose certificate authority %s: %w", resource.Name, err)
		}
	}

	err = createClusterRole(ctx, client, resourceNames)
	if err != nil {
		return fmt.Errorf("failed to create etcd CA admin Role: %w", err)
	}
	return nil
}

func getKubernetesClient(kubeconfigPath string) (*kubernetes.Clientset, error) {
	restConfig, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		return nil, err
	}
	return kubernetes.NewForConfig(restConfig)
}

func createClusterRole(ctx context.Context, client kubernetes.Interface, resourceNames []string) error {
	role := &rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "microshift-ca-admin",
			Namespace: caResourceNamespace,
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups:     []string{""},
				Resources:     []string{"secrets"},
				ResourceNames: resourceNames,
				Verbs:         []string{"*"},
			},
		},
	}

	_, err := client.RbacV1().Roles(caResourceNamespace).Create(ctx, role, metav1.CreateOptions{})
	if err != nil && !apierrors.IsAlreadyExists(err) {
		return fmt.Errorf("failed to create etcd CA admin Role: %w", err)
	}

	roleBinding := &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "microshift-ca-admin-binding",
			Namespace: caResourceNamespace,
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:     "Group",
				Name:     "system:masters",
				APIGroup: rbacv1.GroupName,
			},
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: rbacv1.GroupName,
			Kind:     "Role",
			Name:     "microshift-ca-admin",
		},
	}

	_, err = client.RbacV1().RoleBindings(caResourceNamespace).Create(ctx, roleBinding, metav1.CreateOptions{})
	if err != nil && !apierrors.IsAlreadyExists(err) {
		return fmt.Errorf("failed to create etcd CA admin RoleBinding: %w", err)
	}

	return nil
}

func exposeCertificateAuthority(ctx context.Context, client kubernetes.Interface, dir, name string) error {
	caCertPath := cryptomaterial.CACertPath(dir)
	caKeyPath := cryptomaterial.CAKeyPath(dir)
	serialPath := filepath.Join(dir, "serial.txt")

	caCert, err := os.ReadFile(caCertPath)
	if err != nil {
		return fmt.Errorf("failed to read CA certificate from %s: %w", caCertPath, err)
	}
	caKey, err := os.ReadFile(caKeyPath)
	if err != nil {
		return fmt.Errorf("failed to read CA key from %s: %w", caKeyPath, err)
	}
	serial, err := os.ReadFile(serialPath)
	if err != nil {
		return fmt.Errorf("failed to read CA serial: %w", err)
	}

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: caResourceNamespace,
		},
		Type: corev1.SecretTypeOpaque,
		Data: map[string][]byte{
			"ca.crt":     caCert,
			"ca.key":     caKey,
			"serial.txt": serial,
		},
	}

	caBundlePath := cryptomaterial.CABundlePath(dir)
	if _, err := os.Stat(caBundlePath); err == nil {
		caBundle, err := os.ReadFile(caBundlePath)
		if err != nil {
			return fmt.Errorf("failed to read CA bundle from %s: %w", caBundlePath, err)
		}
		secret.Data["ca-bundle.crt"] = caBundle
	}

	_, err = client.CoreV1().Secrets(caResourceNamespace).Create(ctx, secret, metav1.CreateOptions{})
	if err != nil && !apierrors.IsAlreadyExists(err) {
		return fmt.Errorf("failed to create CA secret %s: %w", name, err)
	}
	return nil
}
