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
	// Resource name for etcd CA certificate
	etcdCAResourceName = "microshift-etcd-ca"
	// Resource namespace
	etcdCAResourceNamespace = "kube-system"
)

func startEtcdController(ctx context.Context, kubeconfigPath string) error {
	client, err := getKubernetesClient(kubeconfigPath)
	if err != nil {
		return fmt.Errorf("failed to get Kubernetes client: %w", err)
	}
	err = exposeEtcdCA(ctx, client)
	if err != nil {
		return fmt.Errorf("failed to expose etcd CA: %w", err)
	}
	err = createClusterRole(ctx, client)
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

func createClusterRole(ctx context.Context, client kubernetes.Interface) error {
	role := &rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "microshift-etcd-ca-admin",
			Namespace: etcdCAResourceNamespace,
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups:     []string{""},
				Resources:     []string{"secrets"},
				ResourceNames: []string{etcdCAResourceName},
				Verbs:         []string{"*"},
			},
		},
	}

	_, err := client.RbacV1().Roles(etcdCAResourceNamespace).Create(ctx, role, metav1.CreateOptions{})
	if err != nil && !apierrors.IsAlreadyExists(err) {
		return fmt.Errorf("failed to create etcd CA admin Role: %w", err)
	}

	roleBinding := &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "microshift-etcd-ca-admin-binding",
			Namespace: etcdCAResourceNamespace,
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
			Name:     "microshift-etcd-ca-admin",
		},
	}

	_, err = client.RbacV1().RoleBindings(etcdCAResourceNamespace).Create(ctx, roleBinding, metav1.CreateOptions{})
	if err != nil && !apierrors.IsAlreadyExists(err) {
		return fmt.Errorf("failed to create etcd CA admin RoleBinding: %w", err)
	}

	return nil
}

func exposeEtcdCA(ctx context.Context, client kubernetes.Interface) error {
	certsDir := cryptomaterial.CertsDirectory(config.DataDir)
	etcdSignerDir := cryptomaterial.EtcdSignerDir(certsDir)
	etcdCACertPath := cryptomaterial.CACertPath(etcdSignerDir)
	etcdCAKeyPath := cryptomaterial.CAKeyPath(etcdSignerDir)

	caCert, err := os.ReadFile(etcdCACertPath)
	if err != nil {
		return fmt.Errorf("failed to read etcd CA certificate from %s: %w", etcdCACertPath, err)
	}

	caKey, err := os.ReadFile(etcdCAKeyPath)
	if err != nil {
		return fmt.Errorf("failed to read etcd CA key from %s: %w", etcdCAKeyPath, err)
	}

	serial, err := os.ReadFile(filepath.Join(etcdSignerDir, "serial.txt"))
	if err != nil {
		return fmt.Errorf("failed to read CA serial: %w", err)
	}

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      etcdCAResourceName,
			Namespace: etcdCAResourceNamespace,
		},
		Type: corev1.SecretTypeOpaque,
		Data: map[string][]byte{
			"ca.crt":     caCert,
			"ca.key":     caKey,
			"serial.txt": serial,
		},
	}

	_, err = client.CoreV1().Secrets(etcdCAResourceNamespace).Create(ctx, secret, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create etcd CA secret: %w", err)
	}
	return nil
}
