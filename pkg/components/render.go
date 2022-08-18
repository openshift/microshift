package components

import (
	"bytes"
	"fmt"
	"path/filepath"
	"text/template"

	"github.com/openshift/microshift/pkg/assets"
	"github.com/openshift/microshift/pkg/config"
	"github.com/openshift/microshift/pkg/config/lvmd"
	"github.com/openshift/microshift/pkg/release"
	"sigs.k8s.io/yaml"
)

var templateFuncs = map[string]interface{}{
	"Dir": filepath.Dir,
}

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

func renderLvmdParams(l *lvmd.Lvmd) (assets.RenderParams, error) {
	r := make(assets.RenderParams)
	b, err := yaml.Marshal(l)
	if err != nil {
		return nil, err
	}
	r["lvmd"] = fmt.Sprintf(`%q`, b)
	return r, nil
}

