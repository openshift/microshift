package components

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"path/filepath"
	"strings"
	"text/template"

	"sigs.k8s.io/yaml"

	"github.com/openshift/microshift/pkg/assets"
	"github.com/openshift/microshift/pkg/config"
	"github.com/openshift/microshift/pkg/config/lvmd"
	"github.com/openshift/microshift/pkg/release"
)

var templateFuncs = map[string]interface{}{
	"Dir":       filepath.Dir,
	"Sha256sum": func(s string) string { return fmt.Sprintf("%x", sha256.Sum256([]byte(s))) },
}

func renderParamsFromConfig(cfg *config.Config, extra assets.RenderParams) assets.RenderParams {
	params := map[string]interface{}{
		"ReleaseImage": release.Image,
		"NodeName":     cfg.CanonicalNodeName(),
		"NodeIP":       cfg.Node.NodeIP,
		"ClusterCIDR":  strings.Join(cfg.Network.ClusterNetwork, ","),
		"ServiceCIDR":  strings.Join(cfg.Network.ServiceNetwork, ","),
		"ClusterDNS":   cfg.Network.DNS,
		"BaseDomain":   cfg.DNS.BaseDomain,
	}
	for k, v := range extra {
		params[k] = v
	}
	return params
}

func renderTemplate(tb []byte, data assets.RenderParams) ([]byte, error) {
	tmpl, err := template.New("").Option("missingkey=error").Funcs(templateFuncs).Parse(string(tb))
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
	content := string(b)
	if l.Message != "" {
		content = fmt.Sprintf("# %s\n%s", l.Message, content)
	}
	r["lvmd"] = content
	r["SocketName"] = l.SocketName
	return r, nil
}
