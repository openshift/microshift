package c2cc

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"text/template"

	"github.com/apparentlymart/go-cidr/cidr"
	"github.com/openshift/microshift/pkg/assets"
	"github.com/openshift/microshift/pkg/release"
	"k8s.io/klog/v2"
)

var (
	c2ccNamespace          = []string{"components/c2cc/namespace.yaml"}
	c2ccServiceAccount     = []string{"components/c2cc/serviceaccount.yaml"}
	c2ccClusterRole        = []string{"components/c2cc/clusterrole.yaml"}
	c2ccClusterRoleBinding = []string{"components/c2cc/clusterrolebinding.yaml"}
	c2ccDeployment         = []string{"components/c2cc/deployment.yaml"}
	c2ccService            = []string{"components/c2cc/service.yaml"}
)

func (c *C2CCRouteManager) deployProbe(ctx context.Context) error {
	var probeIPs []string
	for _, svcNetStr := range c.cfg.Network.ServiceNetwork {
		_, svcNet, err := net.ParseCIDR(svcNetStr)
		if err != nil {
			return fmt.Errorf("failed to parse local service network %q: %w", svcNetStr, err)
		}
		probeIP, err := cidr.Host(svcNet, 11)
		if err != nil {
			return fmt.Errorf("failed to compute probe service ClusterIP from %q: %w", svcNetStr, err)
		}
		probeIPs = append(probeIPs, probeIP.String())
	}

	params := assets.RenderParams{
		"ReleaseImage":          release.Image,
		"ProbeServiceClusterIP": probeIPs[0],
		"DualStack":             len(probeIPs) > 1,
	}
	if len(probeIPs) > 1 {
		params["ProbeServiceClusterIPSecondary"] = probeIPs[1]
	}

	if err := assets.ApplyNamespaces(ctx, c2ccNamespace, c.kubeconfig); err != nil {
		return fmt.Errorf("failed to apply c2cc namespace: %w", err)
	}
	if err := assets.ApplyServiceAccounts(ctx, c2ccServiceAccount, c.kubeconfig); err != nil {
		return fmt.Errorf("failed to apply c2cc service account: %w", err)
	}
	if err := assets.ApplyClusterRoles(ctx, c2ccClusterRole, c.kubeconfig); err != nil {
		return fmt.Errorf("failed to apply c2cc cluster role: %w", err)
	}
	if err := assets.ApplyClusterRoleBindings(ctx, c2ccClusterRoleBinding, c.kubeconfig); err != nil {
		return fmt.Errorf("failed to apply c2cc cluster role binding: %w", err)
	}
	if err := assets.ApplyDeployments(ctx, c2ccDeployment, renderTemplate, params, c.kubeconfig); err != nil {
		return fmt.Errorf("failed to apply c2cc deployment: %w", err)
	}
	if err := assets.ApplyServices(ctx, c2ccService, renderTemplate, params, c.kubeconfig); err != nil {
		return fmt.Errorf("failed to apply c2cc service: %w", err)
	}

	klog.V(4).Infof("C2CC probe assets deployed (probe ClusterIPs=%v)", probeIPs)
	return nil
}

func renderTemplate(tb []byte, data assets.RenderParams) ([]byte, error) {
	tmpl, err := template.New("").Option("missingkey=error").Parse(string(tb))
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
