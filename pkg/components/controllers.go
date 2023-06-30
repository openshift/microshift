package components

import (
	"context"
	"os"

	"github.com/openshift/microshift/pkg/assets"
	"github.com/openshift/microshift/pkg/config"
	"github.com/openshift/microshift/pkg/util/cryptomaterial"
	"k8s.io/klog/v2"
)

func startServiceCAController(ctx context.Context, cfg *config.Config, kubeconfigPath string) error {
	var (
		//TODO: fix the rolebinding and sa
		clusterRoleBinding = []string{
			"components/service-ca/clusterrolebinding.yaml",
		}
		clusterRole = []string{
			"components/service-ca/clusterrole.yaml",
		}
		roleBinding = []string{
			"components/service-ca/rolebinding.yaml",
		}
		role = []string{
			"components/service-ca/role.yaml",
		}
		apps = []string{
			"components/service-ca/deployment.yaml",
		}
		ns = []string{
			"components/service-ca/ns.yaml",
		}
		sa = []string{
			"components/service-ca/sa.yaml",
		}
		secret     = "components/service-ca/signing-secret.yaml"
		secretName = "signing-key"
		cm         = "components/service-ca/signing-cabundle.yaml"
		cmName     = "signing-cabundle"
	)

	serviceCADir := cryptomaterial.ServiceCADir(cryptomaterial.CertsDirectory(config.DataDir))
	caCertPath := cryptomaterial.CACertPath(serviceCADir)
	caKeyPath := cryptomaterial.CAKeyPath(serviceCADir)

	cmData := map[string]string{}
	secretData := map[string][]byte{}

	caCertPEM, err := os.ReadFile(caCertPath)
	if err != nil {
		return err
	}
	caKeyPEM, err := os.ReadFile(caKeyPath)
	if err != nil {
		return err
	}
	cmData["ca-bundle.crt"] = string(caCertPEM)
	secretData["tls.crt"] = caCertPEM
	secretData["tls.key"] = caKeyPEM

	if err := assets.ApplyNamespaces(ctx, ns, kubeconfigPath); err != nil {
		klog.Warningf("Failed to apply ns %v: %v", ns, err)
		return err
	}
	if err := assets.ApplyClusterRoleBindings(ctx, clusterRoleBinding, kubeconfigPath); err != nil {
		klog.Warningf("Failed to apply clusterRolebinding %v: %v", clusterRoleBinding, err)
		return err
	}
	if err := assets.ApplyClusterRoles(ctx, clusterRole, kubeconfigPath); err != nil {
		klog.Warningf("Failed to apply clusterRole %v: %v", clusterRole, err)
		return err
	}
	if err := assets.ApplyRoleBindings(ctx, roleBinding, kubeconfigPath); err != nil {
		klog.Warningf("Failed to apply rolebinding %v: %v", roleBinding, err)
		return err
	}
	if err := assets.ApplyRoles(ctx, role, kubeconfigPath); err != nil {
		klog.Warningf("Failed to apply role %v: %v", role, err)
		return err
	}
	if err := assets.ApplyServiceAccounts(ctx, sa, kubeconfigPath); err != nil {
		klog.Warningf("Failed to apply sa %v: %v", sa, err)
		return err
	}
	if err := assets.ApplySecretWithData(ctx, secret, secretData, kubeconfigPath); err != nil {
		klog.Warningf("Failed to apply secret %v: %v", secret, err)
		return err
	}
	if err := assets.ApplyConfigMapWithData(ctx, cm, cmData, kubeconfigPath); err != nil {
		klog.Warningf("Failed to apply configMap %v: %v", cm, err)
		return err
	}
	extraParams := assets.RenderParams{
		"CAConfigMap": cmName,
		"TLSSecret":   secretName,
	}
	if err := assets.ApplyDeployments(ctx, apps, renderTemplate, renderParamsFromConfig(cfg, extraParams), kubeconfigPath); err != nil {
		klog.Warningf("Failed to apply apps %v: %v", apps, err)
		return err
	}
	return nil
}

func startIngressController(ctx context.Context, cfg *config.Config, kubeconfigPath string) error {
	var (
		clusterRoleBinding = []string{
			"components/openshift-router/cluster-role-binding.yaml",
		}
		clusterRole = []string{
			"components/openshift-router/cluster-role.yaml",
			"components/openshift-router/cluster-role-aggregate-route.yaml",
		}
		apps = []string{
			"components/openshift-router/deployment.yaml",
		}
		ns = []string{
			"components/openshift-router/namespace.yaml",
		}
		sa = []string{
			"components/openshift-router/service-account.yaml",
		}
		svc = []string{
			"components/openshift-router/service-internal.yaml",
		}
		cm                   = "components/openshift-router/configmap.yaml"
		servingKeypairSecret = "components/openshift-router/serving-certificate.yaml"
	)
	if err := assets.ApplyNamespaces(ctx, ns, kubeconfigPath); err != nil {
		klog.Warningf("Failed to apply namespaces %v: %v", ns, err)
		return err
	}
	if err := assets.ApplyClusterRoles(ctx, clusterRole, kubeconfigPath); err != nil {
		klog.Warningf("Failed to apply clusterRole %v: %v", clusterRole, err)
		return err
	}
	if err := assets.ApplyClusterRoleBindings(ctx, clusterRoleBinding, kubeconfigPath); err != nil {
		klog.Warningf("Failed to apply clusterRolebinding %v: %v", clusterRoleBinding, err)
		return err
	}
	if err := assets.ApplyServiceAccounts(ctx, sa, kubeconfigPath); err != nil {
		klog.Warningf("Failed to apply serviceAccount %v %v", sa, err)
		return err
	}

	serviceCADir := cryptomaterial.ServiceCADir(cryptomaterial.CertsDirectory(config.DataDir))
	caCertPath := cryptomaterial.CACertPath(serviceCADir)
	cmData := map[string]string{}

	caCertPEM, err := os.ReadFile(caCertPath)
	if err != nil {
		return err
	}
	cmData["service-ca.crt"] = string(caCertPEM)

	if err := assets.ApplyConfigMapWithData(ctx, cm, cmData, kubeconfigPath); err != nil {
		klog.Warningf("Failed to apply configMap %v: %v", cm, err)
		return err
	}
	if err := assets.ApplyServices(ctx, svc, nil, nil, kubeconfigPath); err != nil {
		klog.Warningf("Failed to apply service %v %v", svc, err)
		return err
	}
	if err := assets.ApplySecretWithData(
		ctx,
		servingKeypairSecret,
		map[string][]byte{
			"tls.crt": cfg.Ingress.ServingCertificate,
			"tls.key": cfg.Ingress.ServingKey,
		},
		kubeconfigPath,
	); err != nil {
		klog.Warningf("failed to apply secret %q: %v", servingKeypairSecret, err)
		return err
	}

	if err := assets.ApplyDeployments(ctx, apps, renderTemplate, renderParamsFromConfig(cfg, nil), kubeconfigPath); err != nil {
		klog.Warningf("Failed to apply apps %v: %v", apps, err)
		return err
	}
	return nil
}

func startDNSController(ctx context.Context, cfg *config.Config, kubeconfigPath string) error {
	var (
		clusterRoleBinding = []string{
			"components/openshift-dns/dns/cluster-role-binding.yaml",
		}
		clusterRole = []string{
			"components/openshift-dns/dns/cluster-role.yaml",
		}
		apps = []string{
			"components/openshift-dns/dns/daemonset.yaml",
			"components/openshift-dns/node-resolver/daemonset.yaml",
		}
		ns = []string{
			"components/openshift-dns/dns/namespace.yaml",
		}
		sa = []string{
			"components/openshift-dns/dns/service-account.yaml",
			"components/openshift-dns/node-resolver/service-account.yaml",
		}
		cm = []string{
			"components/openshift-dns/dns/configmap.yaml",
		}
		svc = []string{
			"components/openshift-dns/dns/service.yaml",
		}
	)
	if err := assets.ApplyNamespaces(ctx, ns, kubeconfigPath); err != nil {
		klog.Warningf("Failed to apply", "namespace", ns, "err", err)
		return err
	}

	extraParams := assets.RenderParams{
		"ClusterIP": cfg.Network.DNS,
	}
	if err := assets.ApplyServices(ctx, svc, renderTemplate, renderParamsFromConfig(cfg, extraParams), kubeconfigPath); err != nil {
		klog.Warningf("Failed to apply service %v %v", svc, err)
		// service already created by coreDNS, not re-create it.
		return nil
	}
	if err := assets.ApplyClusterRoles(ctx, clusterRole, kubeconfigPath); err != nil {
		klog.Warningf("Failed to apply clusterRole %v %v", clusterRole, err)
		return err
	}
	if err := assets.ApplyClusterRoleBindings(ctx, clusterRoleBinding, kubeconfigPath); err != nil {
		klog.Warningf("Failed to apply clusterRoleBinding %v %v", clusterRoleBinding, err)
		return err
	}
	if err := assets.ApplyServiceAccounts(ctx, sa, kubeconfigPath); err != nil {
		klog.Warningf("Failed to apply serviceAccount %v %v", sa, err)
		return err
	}
	// set DNS forward to systemd-resolved resolver if exists
	// https://github.com/coredns/coredns/blob/master/plugin/loop/README.md#troubleshooting-loops-in-kubernetes-clusters
	if _, err := os.Stat(config.DefaultSystemdResolvedFile); err == nil {
		extraParams["UpstreamResolver"] = config.DefaultSystemdResolvedFile
	} else {
		extraParams["UpstreamResolver"] = ""
	}
	if err := assets.ApplyConfigMaps(ctx, cm, renderTemplate, renderParamsFromConfig(cfg, extraParams), kubeconfigPath); err != nil {
		klog.Warningf("Failed to apply configMap %v %v", cm, err)
		return err
	}
	if err := assets.ApplyDaemonSets(ctx, apps, renderTemplate, renderParamsFromConfig(cfg, extraParams), kubeconfigPath); err != nil {
		klog.Warningf("Failed to apply apps %v %v", apps, err)
		return err
	}
	return nil
}
