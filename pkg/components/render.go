package components

import (
	"bytes"
	"text/template"

	"github.com/openshift/microshift/pkg/assets"
	"github.com/openshift/microshift/pkg/config"
	"github.com/openshift/microshift/pkg/release"
)

var templateFuncs = map[string]interface{}{}

func renderParamsFromConfig(cfg *config.MicroshiftConfig, extra assets.RenderParams) assets.RenderParams {
	params := map[string]interface{}{
		"ReleaseImage":  release.Image,
		"NodeName":      cfg.NodeName,
		"NodeIP":        cfg.NodeIP,
		"ClusterCIDR":   cfg.Cluster.ClusterCIDR,
		"ServiceCIDR":   cfg.Cluster.ServiceCIDR,
		"ClusterDNS":    cfg.Cluster.DNS,
		"ClusterDomain": cfg.Cluster.Domain,
		"MTU":           cfg.Cluster.MTU,
	}
	for k, v := range extra {
		params[k] = v
	}
	return params
}

func renderTemplate(tb []byte, data assets.RenderParams) ([]byte, error) {
	tmpl, err := template.New("").Funcs(templateFuncs).Parse(string(tb))
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
