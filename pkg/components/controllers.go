package components

import (
	"os"

	"github.com/openshift/microshift/pkg/assets"
	"github.com/openshift/microshift/pkg/config"
	"k8s.io/klog/v2"
)

func startServiceCAController(cfg *config.MicroshiftConfig, kubeconfigPath string) error {
	var (
		//TODO: fix the rolebinding and sa
		clusterRoleBinding = []string{
			"assets/rbac/0000_60_service-ca_00_clusterrolebinding.yaml",
		}
		clusterRole = []string{
			"assets/rbac/0000_60_service-ca_00_clusterrole.yaml",
		}
		roleBinding = []string{
			"assets/rbac/0000_60_service-ca_00_rolebinding.yaml",
		}
		role = []string{
			"assets/rbac/0000_60_service-ca_00_role.yaml",
		}
		apps = []string{
			"assets/apps/0000_60_service-ca_05_deploy.yaml",
		}
		ns = []string{
			"assets/core/0000_60_service-ca_01_namespace.yaml",
		}
		sa = []string{
			"assets/core/0000_60_service-ca_04_sa.yaml",
		}
		secret     = "assets/core/0000_60_service-ca_04_secret.yaml"
		secretName = "signing-key"
		cm         = "assets/core/0000_60_service-ca_04_configmap.yaml"
		cmName     = "signing-cabundle"
	)
	caPath := cfg.DataDir + "/certs/ca-bundle/ca-bundle.crt"
	tlsCrtPath := cfg.DataDir + "/resources/service-ca/secrets/service-ca/tls.crt"
	tlsKeyPath := cfg.DataDir + "/resources/service-ca/secrets/service-ca/tls.key"
	cmData := map[string]string{}
	secretData := map[string][]byte{}
	cabundle, err := os.ReadFile(caPath)
	if err != nil {
		return err
	}
	tlscrt, err := os.ReadFile(tlsCrtPath)
	if err != nil {
		return err
	}
	tlskey, err := os.ReadFile(tlsKeyPath)
	if err != nil {
		return err
	}
	cmData["ca-bundle.crt"] = string(cabundle)
	secretData["tls.crt"] = tlscrt
	secretData["tls.key"] = tlskey

	if err := assets.ApplyNamespaces(ns, kubeconfigPath); err != nil {
		klog.Warningf("Failed to apply ns %v: %v", ns, err)
		return err
	}
	if err := assets.ApplyClusterRoleBindings(clusterRoleBinding, kubeconfigPath); err != nil {
		klog.Warningf("Failed to apply clusterRolebinding %v: %v", clusterRoleBinding, err)
		return err
	}
	if err := assets.ApplyClusterRoles(clusterRole, kubeconfigPath); err != nil {
		klog.Warningf("Failed to apply clusterRole %v: %v", clusterRole, err)
		return err
	}
	if err := assets.ApplyRoleBindings(roleBinding, kubeconfigPath); err != nil {
		klog.Warningf("Failed to apply rolebinding %v: %v", roleBinding, err)
		return err
	}
	if err := assets.ApplyRoles(role, kubeconfigPath); err != nil {
		klog.Warningf("Failed to apply role %v: %v", role, err)
		return err
	}
	if err := assets.ApplyServiceAccounts(sa, kubeconfigPath); err != nil {
		klog.Warningf("Failed to apply sa %v: %v", sa, err)
		return err
	}
	if err := assets.ApplySecretWithData(secret, secretData, kubeconfigPath); err != nil {
		klog.Warningf("Failed to apply secret %v: %v", secret, err)
		return err
	}
	if err := assets.ApplyConfigMapWithData(cm, cmData, kubeconfigPath); err != nil {
		klog.Warningf("Failed to apply sa %v: %v", cm, err)
		return err
	}
	if err := assets.ApplyDeployments(apps, renderServiceCAController, assets.RenderParams{"ConfigMap": cmName, "Secret": secretName}, kubeconfigPath); err != nil {
		klog.Warningf("Failed to apply apps %v: %v", apps, err)
		return err
	}
	return nil
}

func startIngressController(cfg *config.MicroshiftConfig, kubeconfigPath string) error {
	var (
		clusterRoleBinding = []string{
			"assets/rbac/0000_80_openshift-router-cluster-role-binding.yaml",
		}
		clusterRole = []string{
			"assets/rbac/0000_80_openshift-router-cluster-role.yaml",
		}
		apps = []string{
			"assets/apps/0000_80_openshift-router-deployment.yaml",
		}
		ns = []string{
			"assets/core/0000_80_openshift-router-namespace.yaml",
		}
		sa = []string{
			"assets/core/0000_80_openshift-router-service-account.yaml",
		}
		cm = []string{
			"assets/core/0000_80_openshift-router-cm.yaml",
		}
		svc = []string{
			"assets/core/0000_80_openshift-router-service.yaml",
		}
		extSvc = []string{
			"assets/core/0000_80_openshift-router-external-service.yaml",
		}
	)
	if err := assets.ApplyNamespaces(ns, kubeconfigPath); err != nil {
		klog.Warningf("Failed to apply namespaces %v: %v", ns, err)
		return err
	}
	if err := assets.ApplyClusterRoles(clusterRole, kubeconfigPath); err != nil {
		klog.Warningf("Failed to apply clusterRole %v: %v", clusterRole, err)
		return err
	}
	if err := assets.ApplyClusterRoleBindings(clusterRoleBinding, kubeconfigPath); err != nil {
		klog.Warningf("Failed to apply clusterRolebinding %v: %v", clusterRoleBinding, err)
		return err
	}
	if err := assets.ApplyServiceAccounts(sa, kubeconfigPath); err != nil {
		klog.Warningf("Failed to apply serviceAccount %v %v", sa, err)
		return err
	}
	if err := assets.ApplyConfigMaps(cm, kubeconfigPath); err != nil {
		klog.Warningf("Failed to apply configMap %v, %v", cm, err)
		return err
	}
	if err := assets.ApplyServices(svc, nil, nil, kubeconfigPath); err != nil {
		klog.Warningf("Failed to apply service %v %v", svc, err)
		return err
	}
	if err := assets.ApplyServices(extSvc, nil, nil, kubeconfigPath); err != nil {
		klog.Warningf("Failed to apply external ingress svc %v: %v", extSvc, err)
		return err
	}
	if err := assets.ApplyDeployments(apps, renderReleaseImage, nil, kubeconfigPath); err != nil {
		klog.Warningf("Failed to apply apps %v: %v", apps, err)
		return err
	}
	return nil
}

func startDNSController(cfg *config.MicroshiftConfig, kubeconfigPath string) error {
	var (
		clusterRoleBinding = []string{
			"assets/rbac/0000_70_dns_01-cluster-role-binding.yaml",
		}
		clusterRole = []string{
			"assets/rbac/0000_70_dns_01-cluster-role.yaml",
		}
		apps = []string{
			"assets/apps/0000_70_dns_01-dns-daemonset.yaml",
			"assets/apps/0000_70_dns_01-node-resolver-daemonset.yaml",
		}
		ns = []string{
			"assets/core/0000_70_dns_00-namespace.yaml",
		}
		sa = []string{
			"assets/core/0000_70_dns_01-dns-service-account.yaml",
			"assets/core/0000_70_dns_01-node-resolver-service-account.yaml",
		}
		cm = []string{
			"assets/core/0000_70_dns_01-configmap.yaml",
		}
		svc = []string{
			"assets/core/0000_70_dns_01-service.yaml",
		}
	)
	if err := assets.ApplyNamespaces(ns, kubeconfigPath); err != nil {
		klog.Warningf("Failed to apply", "namespace", ns, "err", err)
		return err
	}
	if err := assets.ApplyServices(svc, renderDNSService, assets.RenderParams{"ClusterDNS": cfg.Cluster.DNS}, kubeconfigPath); err != nil {
		klog.Warningf("Failed to apply service %v %v", svc, err)
		// service already created by coreDNS, not re-create it.
		return nil
	}
	if err := assets.ApplyClusterRoles(clusterRole, kubeconfigPath); err != nil {
		klog.Warningf("Failed to apply clusterRole %v %v", clusterRole, err)
		return err
	}
	if err := assets.ApplyClusterRoleBindings(clusterRoleBinding, kubeconfigPath); err != nil {
		klog.Warningf("Failed to apply clusterRoleBinding %v %v", clusterRoleBinding, err)
		return err
	}
	if err := assets.ApplyServiceAccounts(sa, kubeconfigPath); err != nil {
		klog.Warningf("Failed to apply serviceAccount %v %v", sa, err)
		return err
	}
	if err := assets.ApplyConfigMaps(cm, kubeconfigPath); err != nil {
		klog.Warningf("Failed to apply configMap %v %v", cm, err)
		return err
	}
	if err := assets.ApplyDaemonSets(apps, renderReleaseImage, nil, kubeconfigPath); err != nil {
		klog.Warningf("Failed to apply apps %v %v", apps, err)
		return err
	}
	return nil
}
